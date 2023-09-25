package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	grpcPeer "github.com/birros/go-libp2p-grpc"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	codes "google.golang.org/grpc/codes"
)

var chatPasswd string = "password"

// Peer informantion
type Peer struct {
	id      string
	handler *gs.ChatRoom_ListenServer
}

type ChatServer struct {
	gs.UnimplementedChatRoomServer

	registeredAccount gs.UserList
	connectedPeers    map[string]Peer

	mu sync.Mutex
}

// ---------------------------------------------------------//
// ------------------ HELPER -------------------------------//

// Check if a peer already connect to server
func (cs *ChatServer) isConnected(username string) bool {
	if _, ok := cs.connectedPeers[username]; !ok {
		return false
	}

	return true
}

// Load credentials from a json file
func (cs *ChatServer) loadUserInformation(filePath string) error {
	// Read the JSON file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Unmarshal the JSON data into UserList
	if err := json.Unmarshal(jsonData, &cs.registeredAccount); err != nil {
		return err
	}

	return nil
}

// ---------------------------------------------------------//

// ------------------------- RPC ---------------------------//

func (cs *ChatServer) SendPublicMessage(ctx context.Context, msg *gs.ChatMessage) (*gs.SentMessageStatus, error) {
	log.Printf("Recieved message from %s: %s\n", msg.Sender, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	cs.mu.Lock()
	for username, peer := range cs.connectedPeers {
		if username == msg.GetSender() {
			continue
		}

		log.Printf("Broadcast message to %s\n", username)
		handler := *peer.handler
		if err := handler.Send(msg); err != nil {
			log.Printf("Failed to send message to %s: %s", username, err.Error())
		}
	}
	cs.mu.Unlock()

	log.Println("Boradcast message completed!")

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		nil
}

func (cs *ChatServer) SendPrivateMessage(ctx context.Context, msg *gs.PrivateChatMessage) (*gs.SentMessageStatus, error) {
	log.Printf("%s sent a message to %s: %s\n", msg.Sender, msg.Recipent, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	handler := *(cs.connectedPeers[msg.Recipent].handler)
	handler.Send(&gs.ChatMessage{
		Sender:  msg.GetSender(),
		Message: msg.GetMessage(),
	})

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		nil
}

func (cs *ChatServer) Listen(cmd *gs.Command, handler gs.ChatRoom_ListenServer) error {
	username := cmd.GetAdditionalInfo()

	if len(username) == 0 {
		return errors.New("Blank username is not allowed")
	}

	if cs.connectedPeers[username].handler != nil {
		return errors.New("User already connected from another place. Try again later")
	}

	peer := cs.connectedPeers[username]
	peer.handler = &handler
	cs.connectedPeers[username] = peer

	return nil
}

// Login to server using registered account (username and password)
func (cs *ChatServer) Login(ctx context.Context, in *gs.UserLoginCredentials) (*gs.AuthenticationResult, error) {
	peerId, isGood := grpcPeer.RemotePeerFromContext(ctx)

	result := gs.AuthenticationResult{}
	result.Username = in.Username
	result.Status = int32(codes.Unauthenticated)

	// cannot get peer information
	if !isGood {
		log.Fatalf("Cannot authenticate the user %s\n", in.Username)
		*result.Message = "Failed to authenticate credentials!"
		result.Status = int32(codes.Unavailable)

		return &result, nil
	}

	if len(cs.connectedPeers) == 0 {
		cs.connectedPeers = make(map[string]Peer)
	} else if _, ok := cs.connectedPeers[in.GetUsername()]; ok {
		*result.Message = fmt.Sprintf("Failed to login as %s: Already login from another place!", in.Username)
		return &result, nil
	} else if in.Password != chatPasswd {
		*result.Message = fmt.Sprintf("Failed to login as %s: Wrong password!", in.Username)
	}

	cs.connectedPeers[in.Username] = Peer{
		id:      string(peerId),
		handler: nil,
	}
	*result.Message = "Login successfully!!!!"
	result.Status = int32(codes.OK)

	return &result, nil
}

// ---------------------------------------------------------//
