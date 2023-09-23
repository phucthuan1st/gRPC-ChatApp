package p2pgrpc

import (
	"context"
	"io"
	"net"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	mnet "github.com/multiformats/go-multiaddr/net"
)

var _ net.Listener = (*listener)(nil)

type listener struct {
	h        host.Host
	streamCh chan network.Stream
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewListener turns a libp2p host into a net listener.
func NewListener(
	ctx context.Context,
	h host.Host,
	pid protocol.ID,
) net.Listener {
	l := listener{
		h:        h,
		streamCh: make(chan network.Stream),
	}
	l.ctx, l.cancel = context.WithCancel(ctx)

	h.SetStreamHandler(pid, func(s network.Stream) {
		l.streamCh <- s
	})

	return l
}

func (l listener) Accept() (net.Conn, error) {
	select {
	case <-l.ctx.Done():
		return nil, io.EOF
	case s := <-l.streamCh:
		return &Conn{Stream: s}, nil
	}
}

func (l listener) Addr() net.Addr {
	addrs := l.h.Network().ListenAddresses()
	if len(addrs) > 0 {
		for _, a := range addrs {
			na, err := mnet.ToNetAddr(a)
			if err == nil {
				return na
			}
		}
	}

	return fakeAddr()
}

func (l listener) Close() error {
	l.cancel()
	return nil
}
