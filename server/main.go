package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	be "github.com/phucthuan1st/gRPC-ChatRoom/server/backend"
	"google.golang.org/grpc"
)

func main() {
	// https://grpc.io/docs/languages/go/basics/#starting-the-server
	const port = 55555

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	if err != nil {
		log.Fatalf("Cannot start the server: %s", err.Error())
		return
	}

	grpcServer := grpc.NewServer()
	backendServer := be.NewChatServer(os.Args[1], os.Args[2])
	grpcService.RegisterChatRoomServer(grpcServer, backendServer)

	// Start the gRPC server
	log.Printf("Starting gRPC server on localhost:%d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
