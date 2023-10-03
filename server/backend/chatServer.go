package backend

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

type ChatServer struct {
	registeredAccount gs.UserList
	loggedInAccount   map[string]bool
	clientStream      map[string]gs.ChatRoom_ChatServer
	serverPassword    string
	mu                sync.Mutex
	gs.UnimplementedChatRoomServer
}

// ---------------------------------------------------------//
// ------------------ HELPER -------------------------------//
func (cs *ChatServer) getClientStream(username string) gs.ChatRoom_ChatServer {
	stream, ok := cs.clientStream[username]
	if ok {
		return stream
	} else {
		return nil
	}
}

// Check if a peer already connect to server
func (cs *ChatServer) isConnected(username string) bool {
	stream, ok := cs.clientStream[username]
	return ok && stream != nil
}

func (cs *ChatServer) isLoggedIn(username string) bool {
	_, ok := cs.loggedInAccount[username]
	return ok
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

func (cs *ChatServer) addClientStream(username string, stream gs.ChatRoom_ChatServer) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clientStream[username] = stream
}

func (cs *ChatServer) removeClientStream(username string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.clientStream, username)
}

// ---------------------------------------------------------//
// ------------------------- RPC ---------------------------//

// monitoring chat activities
func (cs *ChatServer) Chat(stream gs.ChatRoom_ChatServer) error {

	/*
		first message received from client will be used as username
		and msg will be taken as validtion token (genrated when register)
	*/
	msg, err := stream.Recv()
	if err != nil {
		log.Fatalf("Error reciving message: %v", err)
		return err
	}
	username := msg.GetSender()
	if !cs.isLoggedIn(username) {
		log.Fatalf("Unlogged in user: %s is not permit to chat!!!\n", username)
		return errors.New("Unlogged in user: " + username + " is not permit to chat!!!")
	}

	token := msg.GetMessage()
	log.Printf("User %s joined the chat room with token %s!\n", username, token)

	// TODO: validate token (later)

	cs.addClientStream(username, stream)
	defer func() {
		log.Printf("Client disconnected: %s!\n", username)
		cs.removeClientStream(username)
	}()

	/*
		other messages received from client will be used as messages in chat
	*/
	for {
		msg, err = stream.Recv()
		if err != nil {
			log.Fatalf("Error reciving message: %v", err)
			break
		}
		log.Printf("Room chat from %s: %s\n", msg.GetSender(), msg.GetMessage())

		for recipient, recvStream := range cs.clientStream {
			if recipient != msg.GetSender() {
				err := recvStream.Send(&gs.ChatMessage{
					Message: msg.GetMessage(),
					Sender:  msg.GetSender(),
				})

				if err != nil {
					log.Fatalf("Error sending message to %s %v\n", recipient, err)
					break
				}
			}
		}
	}

	return err
}

func (cs *ChatServer) SendPrivateMessage(ctx context.Context, msg *gs.PrivateChatMessage) (*gs.SentMessageStatus, error) {

	log.Printf("%s sent a message to %s: %s\n", msg.Sender, msg.Recipent, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	stream, ok := cs.clientStream[msg.Recipent]
	if !ok {
		return nil, errors.New(fmt.Sprintf("User %s not found!", msg.Recipent))
	}

	err := stream.Send(&gs.ChatMessage{
		Sender:  msg.GetSender(),
		Message: msg.GetMessage(),
	})

	return &gs.SentMessageStatus{
			Id:        id,
			Timestamp: timestamp,
			Status:    int32(status)},
		err
}

// Login to server using registered account (username and password)
func (cs *ChatServer) Login(ctx context.Context, in *gs.UserLoginCredentials) (*gs.AuthenticationResult, error) {

	result := gs.AuthenticationResult{}
	result.Username = in.Username
	result.Status = int32(codes.Unauthenticated)

	if cs.isLoggedIn(in.Username) {
		msg := fmt.Sprintf("Failed to login as %s: Already login from another place!", in.Username)
		result.Message = &msg
		log.Fatalf("Failed to login as %s: Already login from another place!", in.Username)
		return &result, nil
	} else if in.Password != cs.serverPassword {
		msg := fmt.Sprintf("Failed to login as %s: Wrong password!", in.Username)
		result.Message = &msg
		log.Fatalf("Failed to login as %s: Wrong password!", in.Username)
		return &result, nil
	}

	msg := "Login successfully!!!!"
	result.Message = &msg
	log.Printf("%s was just logged in\n", in.Username)
	result.Status = int32(codes.OK)
	cs.loggedInAccount[in.Username] = true

	return &result, nil
}

// ---------------------------------------------------------//

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func NewChatServer(serverPassword, pathToUserCredentials string) *ChatServer {
	cs := ChatServer{}
	cs.clientStream = make(map[string]gs.ChatRoom_ChatServer)
	cs.loggedInAccount = make(map[string]bool)
	cs.mu = sync.Mutex{}
	cs.serverPassword = serverPassword
	cs.loadUserInformation(pathToUserCredentials)

	return &cs
}
