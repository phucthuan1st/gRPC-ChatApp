package p2pgrpc

import "net"

func fakeAddr() net.Addr {
	localIP := net.ParseIP("127.0.0.1")
	return &net.TCPAddr{IP: localIP, Port: 0}
}
