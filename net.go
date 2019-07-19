package neo

import (
	"net"
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
}

func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	pp := <-c.packets
	n = copy(p, pp.buf)
	return n, pp.addr, nil
}

func (c *PacketConn) WriteTo(p []byte, a net.Addr) (n int, err error) {
	c.net.peers[a.Network()+"/"+a.String()].packets <- packet{
		addr: c.addr,
		buf:  append([]byte{}, p...),
	}
	return len(p), nil
}

func (c PacketConn) LocalAddr() net.Addr { return c.addr }

func (PacketConn) Close() error                       { panic("implement me") }
func (PacketConn) SetDeadline(t time.Time) error      { panic("implement me") }
func (PacketConn) SetReadDeadline(t time.Time) error  { panic("implement me") }
func (PacketConn) SetWriteDeadline(t time.Time) error { panic("implement me") }

type NetAddr struct {
	Net     string
	Address string
}

func (n NetAddr) Network() string { return n.Net }
func (n NetAddr) String() string  { return n.Address }

func (n *Net) ListenPacket(network, address string) (net.PacketConn, error) {
	a := &NetAddr{
		// TODO: use virtual DNS.
		Address: address,
		Net:     network,
	}
	pc := &PacketConn{
		net:     n,
		addr:    a,
		packets: make(chan packet, 10),
	}
	n.peers[a.Network()+"/"+a.String()] = pc
	return pc, nil
}
