package neo

import (
	"bytes"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestNet_ListenPacket(t *testing.T) {
	nt := &Net{
		peers: make(map[string]*PacketConn),
	}
	left, err := nt.ListenPacket("udp", "10.0.0.1:123")
	if err != nil {
		t.Fatal(err)
	}
	right, err := nt.ListenPacket("udp", "10.0.0.2:123")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = right.WriteTo([]byte("hello world"), left.LocalAddr()); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 1024)
	n, addr, err := left.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}
	if addr.String() != "10.0.0.2:123" {
		t.Errorf("bad addr: %s", addr)
	}
	if string(buf[:n]) != "hello world" {
		t.Errorf("bad message: %s", buf[:n])
	}
}

type pingServer struct {
	conn net.PacketConn
}

func (s *pingServer) Listen() error {
	buf := make([]byte, 1024)
	for {
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return err
		}
		if _, err = s.conn.WriteTo(buf[:n], addr); err != nil {
			return err
		}
	}
}

type pingClient struct {
	conn net.PacketConn
}

func (c *pingClient) Ping(addr net.Addr) error {
	msg := []byte("hello")
	if _, err := c.conn.WriteTo(msg, addr); err != nil {
		return err
	}
	buf := make([]byte, 1024)
	n, rAddr, err := c.conn.ReadFrom(buf)
	if err != nil {
		return err
	}
	if rAddr.String() != addr.String() {
		return errors.New("bad addr")
	}
	if !bytes.Equal(buf[:n], msg) {
		return errors.New("mismatch")
	}
	return nil
}

func TestNetPing(t *testing.T) {
	nt := &Net{
		peers: make(map[string]*PacketConn),
	}
	left, err := nt.ListenPacket("udp", "10.0.0.1:123")
	if err != nil {
		t.Fatal(err)
	}
	right, err := nt.ListenPacket("udp", "10.0.0.2:123")
	if err != nil {
		t.Fatal(err)
	}
	s := &pingServer{conn: left}
	finished := make(chan struct{})
	go func() {
		if listenErr := s.Listen(); listenErr != nil {
			// Wait for close.
			if e, ok := listenErr.(syscall.Errno); ok {
				if e != syscall.EINVAL {
					t.Error(e)
				}
			} else {
				t.Error(listenErr)
			}
		}
		finished <- struct{}{}
	}()

	c := &pingClient{conn: right}
	for i := 0; i < 10; i++ {
		if err = c.Ping(left.LocalAddr()); err != nil {
			t.Fatal(err)
		}
	}
	if err = right.Close(); err != nil {
		t.Fatal(err)
	}
	if err = left.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
		// OK
	case <-time.After(time.Second * 10):
		t.Error("timed out")
	}
}

func TestNetPingDeadline(t *testing.T) {
	t.Skip("skipping flaky test, see https://github.com/gotd/neo/pull/13#issuecomment-1001285136")

	nt := &Net{
		peers: make(map[string]*PacketConn),
	}
	left, err := nt.ListenPacket("udp", "10.0.0.1:123")
	if err != nil {
		t.Fatal(err)
	}
	right, err := nt.ListenPacket("udp", "10.0.0.2:123")
	if err != nil {
		t.Fatal(err)
	}
	s := &pingServer{conn: left}
	finished := make(chan struct{})
	go func() {
		if listenErr := s.Listen(); listenErr != nil {
			// Wait for close.
			if e, ok := listenErr.(syscall.Errno); ok {
				if e != syscall.EINVAL {
					t.Error(e)
				}
			} else {
				t.Error(listenErr)
			}
		}
		finished <- struct{}{}
	}()

	c := &pingClient{conn: right}
	if err = right.SetReadDeadline(time.Now()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 100)
	if err = c.Ping(left.LocalAddr()); err == nil {
		t.Error("should error!")
	}

	if err = right.Close(); err != nil {
		t.Fatal(err)
	}
	if err = left.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
		// OK
	case <-time.After(time.Second * 10):
		t.Error("timed out")
	}
}
