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

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	codes "google.golang.org/grpc/codes"
)

var chatPasswd string = "password"

// Peer informantion
type Peer struct {
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

	var count int = 0

	cs.mu.Lock()
	for username, peer := range cs.connectedPeers {
		if username == msg.GetSender() {
			continue
		}

		log.Printf("Broadcast message to %s\n", username)
		handler := *peer.handler
		if err := handler.Send(msg); err != nil {
			log.Printf("Failed to send message to %s: %s", username, err.Error())
			count++
		}
	}
	cs.mu.Unlock()

	log.Println("Boradcast message completed!")

	var err error = nil

	if count > 0 {
		err = errors.New(fmt.Sprintf("%d users can't receive the message!", count))
	}

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		err
}

func (cs *ChatServer) SendPrivateMessage(ctx context.Context, msg *gs.PrivateChatMessage) (*gs.SentMessageStatus, error) {
	log.Printf("%s sent a message to %s: %s\n", msg.Sender, msg.Recipent, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	handler := *(cs.connectedPeers[msg.Recipent].handler)
	err := handler.Send(&gs.ChatMessage{
		Sender:  msg.GetSender(),
		Message: msg.GetMessage(),
	})

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		err
}

func (cs *ChatServer) Listen(cmd *gs.Command, handler gs.ChatRoom_ListenServer) error {
	username := cmd.GetAdditionalInfo()

	if len(username) == 0 {
		log.Fatalf("A blank named peer was try to join the chat")
		return errors.New("Blank username is not allowed")
	}

	if cs.connectedPeers[username].handler != nil {
		log.Fatalf("User %s already connected from another place. Try again later\n", *cmd.AdditionalInfo)
		return errors.New("User already connected from another place. Try again later")
	}

	peer := cs.connectedPeers[username]
	peer.handler = &handler
	cs.connectedPeers[username] = peer

	return nil
}

// Login to server using registered account (username and password)
func (cs *ChatServer) Login(ctx context.Context, in *gs.UserLoginCredentials) (*gs.AuthenticationResult, error) {

	result := gs.AuthenticationResult{}
	result.Username = in.Username
	result.Status = int32(codes.Unauthenticated)

	if len(cs.connectedPeers) == 0 {
		cs.connectedPeers = make(map[string]Peer)
	} else if _, ok := cs.connectedPeers[in.GetUsername()]; ok {
		msg := fmt.Sprintf("Failed to login as %s: Already login from another place!", in.Username)
		result.Message = &msg
		log.Fatalf("Failed to login as %s: Already login from another place!", in.Username)
		return &result, nil

	} else if in.Password != chatPasswd {

		// TODO: implement login logic
		msg := fmt.Sprintf("Failed to login as %s: Wrong password!", in.Username)
		result.Message = &msg
		log.Fatalf("Failed to login as %s: Wrong password!", in.Username)
		return &result, nil
	}

	cs.connectedPeers[in.Username] = Peer{
		handler: nil,
	}
	msg := "Login successfully!!!!"
	result.Message = &msg
	log.Printf("%s was just logged in\n", in.Username)
	result.Status = int32(codes.OK)

	return &result, nil
}

// ---------------------------------------------------------//
