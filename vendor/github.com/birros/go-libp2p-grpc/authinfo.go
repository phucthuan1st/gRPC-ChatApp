package p2pgrpc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"google.golang.org/grpc/credentials"
	grpcpeer "google.golang.org/grpc/peer"
)

var _ credentials.AuthInfo = (*AuthInfo)(nil)

// AuthInfo embed original libp2p stream
type AuthInfo struct {
	Stream network.Stream
}

// AuthType ...
func (ai AuthInfo) AuthType() string {
	return ""
}

// AuthInfoFromContext extracts p2p AuthInfo from ctx
func AuthInfoFromContext(ctx context.Context) (AuthInfo, bool) {
	p, ok := grpcpeer.FromContext(ctx)
	if !ok {
		return AuthInfo{}, false
	}

	ai, ok := p.AuthInfo.(AuthInfo)
	return ai, ok
}

// RemotePeerFromContext extracts remote peer from ctx
func RemotePeerFromContext(ctx context.Context) (peer.ID, bool) {
	ai, ok := AuthInfoFromContext(ctx)
	if !ok {
		return "", false
	}

	peerID := ai.Stream.Conn().RemotePeer()
	return peerID, true
}
