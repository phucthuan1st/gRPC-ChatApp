package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"

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

	stream, err := client.JoinChat(context.Background())

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
			log.Printf("Got message %s at %d from server, status: %d", in.GetId(), in.GetTimestamp(), in.GetStatus())
		}
	}()

	msg := &gs.ChatMessage{
		Sender:  "Alice",
		Message: ""}

	reader := bufio.NewReader(os.Stdin)

	for msg.GetMessage() != "quit\n" {
		fmt.Print("Enter a message: ")
		msg.Message, _ = reader.ReadString('\n')

		if msg.GetMessage() != "\n" {
			err = stream.Send(msg)

			if err != nil {
				log.Fatalln("failed to send message to server")
			}
		}

		msg.Message = ""
	}

	stream.CloseSend()
	<-waitc
}
