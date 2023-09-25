package main

import (
	"context"
	"fmt"
	"io"
	"log"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// https://grpc.io/docs/languages/go/basics/#client
	const port = 55555
	const ipaddr = "localhost"

	serverAddr := fmt.Sprintf("%s:%d", ipaddr, port)
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Cannot connect to server: %s", err.Error())
		return
	}
	defer conn.Close()

	client := gs.NewChatRoomClient(conn)

	username := "Alice"
	stream, err := client.Listen(context.Background(), &gs.Command{AdditionalInfo: &username})

	// start to wait for server message
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("%s just chat: %s", in.GetSender(), in.GetMessage())
		}
	}()
	<-waitc
}
