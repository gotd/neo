package neo

import "testing"

func TestNet_ListenPacket(t *testing.T) {
	nt := &Net{
		peers: make(map[string]*PacketConn),
	}
	left, err := nt.ListenPacket("v", "foo:123")
	if err != nil {
		t.Fatal(err)
	}
	right, err := nt.ListenPacket("v", "bar:123")
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
	if addr.String() != "bar:123" {
		t.Errorf("bad addr: %s", addr)
	}
	if string(buf[:n]) != "hello world" {
		t.Errorf("bad message: %s", buf[:n])
	}
}
