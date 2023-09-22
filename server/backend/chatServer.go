package backend

import (
	"context"
	"fmt"
	"time"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
)

type ChatServer struct {
	gs.UnimplementedChatRoomServer
	// connectedClients []*gs.ChatRoomClient

	// mu sync.Mutex
}

func (cs *ChatServer) SendMessage(ctx context.Context, msg *gs.ChatMessage) (*gs.SentMessageStatus, error) {
	fmt.Printf("Recieved message from %s: %s\n", msg.Sender, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		nil
}

func (cs *ChatServer) ReceiveMessage(stream gs.ChatRoom_ReceiveMessageServer) error {
	for {
		msg, err := stream.Recv()

		if err != nil {
			return err
		}

		fmt.Printf("Recieved message from %s: %s\n", msg.Sender, msg.Message)

		timestamp := time.Now().Unix()
		id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
		status := 1

		response := &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status),
		}

		err = stream.Send(response)

		if err != nil {
			fmt.Printf("Failed to response to client %s: %s", msg.GetSender(), err.Error())
			return err
		}
	}
}
