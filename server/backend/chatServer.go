package backend

import (
	"context"
	"fmt"
	"log"
	"time"

	grpcPeer "github.com/birros/go-libp2p-grpc"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
)

var chatPasswd string = "password"

// Define status codes as constants
const (
	StatusOK          = 200
	StatusFailed      = 401
	StatusServerError = 500
)

type Peer struct {
	name   string
	peerId string
}

type ChatServer struct {
	gs.UnimplementedChatRoomServer
	connectedClients []Peer

	// mu sync.Mutex
}

func (cs *ChatServer) Contains(peer Peer) bool {
	for _, connectedPeer := range cs.connectedClients {
		if connectedPeer == peer {
			return true
		}
	}

	return false
}

func (cs *ChatServer) Login(ctx context.Context, in *gs.UserCredentials) (*gs.AuthenticationResult, error) {
	peerId, isGood := grpcPeer.RemotePeerFromContext(ctx)

	result := gs.AuthenticationResult{}
	result.Username = in.Username

	if !isGood {
		log.Fatalf("Cannot authenticate the user %s\n", in.Username)
		result.Message = "Failed to authenticate"
		result.Status = StatusServerError

		return &result, nil
	}

	peer := Peer{name: in.Username, peerId: string(peerId)}

	if len(cs.connectedClients) == 0 {
		cs.connectedClients = make([]Peer, 0)
	} else if cs.Contains(peer) == true || in.Password != chatPasswd {
		result.Message = fmt.Sprintf("Already exists the username: %s", in.Username)
		result.Status = StatusFailed

		return &result, nil
	}

	cs.connectedClients = append(cs.connectedClients, peer)
	result.Message = "Login successfully!!!!"
	result.Status = StatusOK

	return &result, nil
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

func (cs *ChatServer) JoinChat(stream gs.ChatRoom_JoinChatServer) error {
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
