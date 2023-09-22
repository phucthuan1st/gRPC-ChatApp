package main

import (
	"context"
	"fmt"
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

	msg := &gs.ChatMessage{
		Sender:  "Alice",
		Message: "Hello Server, I'm here to test the server"}

	response, err := client.SendMessage(context.Background(), msg)

	if err != nil {
		log.Fatalln("Cannot send message to server")
		return
	}

	fmt.Printf("Send complete: %v\n", response)
}
