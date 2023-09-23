package p2pgrpc

import (
	"net"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
	mnet "github.com/multiformats/go-multiaddr/net"
)

var _ net.Conn = (*Conn)(nil)

// Conn as libp2p stream
type Conn struct {
	network.Stream
}

// LocalAddr shows local peer net addr
func (c Conn) LocalAddr() net.Addr {
	a := c.Stream.Conn().LocalMultiaddr()
	return toNetAddr(a)
}

// RemoteAddr shows remote peer net addr
func (c Conn) RemoteAddr() net.Addr {
	a := c.Stream.Conn().RemoteMultiaddr()
	return toNetAddr(a)
}

func toNetAddr(ma multiaddr.Multiaddr) net.Addr {
	na, err := mnet.ToNetAddr(ma)
	if err != nil {
		return fakeAddr()
	}
	return na
}
