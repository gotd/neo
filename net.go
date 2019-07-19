package neo

import (
	"errors"
	"net"
	"strconv"
	"syscall"
	"time"
)

// Net is virtual "net" package, implements mesh of peers.
type Net struct {
	peers map[string]*PacketConn
}

type packet struct {
	buf  []byte
	addr net.Addr
}

// PacketConn simulates mesh peer of Net.
type PacketConn struct {
	packets chan packet
	addr    net.Addr
	net     *Net

	closed bool
}

func addrKey(a net.Addr) string {
	if u, ok := a.(*net.UDPAddr); ok {
		return "udp/" + u.String()
	}
	return a.Network() + "/" + a.String()
}

func (c *PacketConn) ok() bool {
	if c == nil {
		return false
	}
	return !c.closed
}

// ReadFrom reads a packet from the connection,
// copying the payload into p.
func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	if !c.ok() {
		return 0, nil, syscall.EINVAL
	}
	pp := <-c.packets
	n = copy(p, pp.buf)
	return n, pp.addr, nil
}

// WriteTo writes a packet with payload p to addr.
func (c *PacketConn) WriteTo(p []byte, a net.Addr) (n int, err error) {
	if !c.ok() {
		return 0, syscall.EINVAL
	}
	c.net.peers[addrKey(a)].packets <- packet{
		addr: c.addr,
		buf:  append([]byte{}, p...),
	}
	return len(p), nil
}

func (c PacketConn) LocalAddr() net.Addr { return c.addr }

// Close closes the connection.
func (c *PacketConn) Close() error {
	if !c.ok() {
		return syscall.EINVAL
	}
	c.closed = true
	close(c.packets)
	return nil
}

func (PacketConn) SetDeadline(t time.Time) error      { panic("implement me") }
func (PacketConn) SetReadDeadline(t time.Time) error  { panic("implement me") }
func (PacketConn) SetWriteDeadline(t time.Time) error { panic("implement me") }

type NetAddr struct {
	Net     string
	Address string
}

func (n NetAddr) Network() string { return n.Net }
func (n NetAddr) String() string  { return n.Address }

// ResolveUDPAddr returns an address of UDP end point.
func (n *Net) ResolveUDPAddr(network, address string) (*net.UDPAddr, error) {
	a := &net.UDPAddr{
		Port: 0,
		IP:   net.IPv4(127, 0, 0, 1),
	}
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	if a.IP = net.ParseIP(host); a.IP == nil {
		// Probably we should use virtual DNS here.
		return nil, errors.New("bad IP")
	}
	if a.Port, err = strconv.Atoi(port); err != nil {
		return nil, err
	}
	return a, nil
}

// ListenPacket announces on the local network address.
func (n *Net) ListenPacket(network, address string) (net.PacketConn, error) {
	if network != "udp4" && network != "udp" && network != "udp6" {
		return nil, errors.New("bad net")
	}
	a, err := n.ResolveUDPAddr(network, address)
	if err != nil {
		return nil, err
	}
	pc := &PacketConn{
		net:     n,
		addr:    a,
		packets: make(chan packet, 10),
	}
	n.peers[addrKey(a)] = pc
	return pc, nil
}
