package p2pgrpc

import (
	"context"
	"errors"
	"net"

	"github.com/libp2p/go-libp2p/core/network"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var _ credentials.TransportCredentials = (*p2pCredentials)(nil)

type p2pCredentials struct{}

// WithP2PCredentials exposes the libp2p stream to the grpc service via a custom
// AuthInfo.
func WithP2PCredentials() grpc.ServerOption {
	return grpc.Creds(p2pCredentials{})
}

func (pc p2pCredentials) ClientHandshake(
	_ context.Context,
	_ string,
	c net.Conn,
) (net.Conn, credentials.AuthInfo, error) {
	return c, nil, nil
}

func (pc p2pCredentials) ServerHandshake(
	c net.Conn,
) (net.Conn, credentials.AuthInfo, error) {
	s, ok := c.(network.Stream)
	if !ok {
		return nil, nil, errors.New("not a libp2p stream")
	}

	i := AuthInfo{
		Stream: s,
	}
	return c, i, nil
}

func (pc p2pCredentials) Info() credentials.ProtocolInfo {
	return credentials.ProtocolInfo{
		ProtocolVersion:  "",
		SecurityProtocol: "",
		SecurityVersion:  "",
		ServerName:       "",
	}
}

func (pc p2pCredentials) Clone() credentials.TransportCredentials {
	return p2pCredentials{}
}

func (pc p2pCredentials) OverrideServerName(string) error {
	return nil
}
