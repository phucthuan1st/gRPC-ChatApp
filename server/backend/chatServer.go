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
	"strings"
	"sync"
	"time"

	gs "github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChatServer struct {
	registeredAccount     gs.UserList
	loggedInAccount       map[string]bool
	clientStream          map[string]gs.ChatRoom_ChatServer
	messageLikes          map[string]int
	serverPassword        string
	mu                    sync.Mutex
	pathToUserCredentials string
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
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if err := json.Unmarshal(jsonData, &cs.registeredAccount); err != nil {
		return err
	}

	return nil
}

func (cs *ChatServer) writeUserInformation(filePath string) error {
	// Marshal the registeredAccount struct to JSON
	cs.mu.Lock()
	jsonData, err := json.MarshalIndent(&cs.registeredAccount, "", "  ")
	cs.mu.Unlock()
	if err != nil {
		return err
	}

	// Write the JSON data to the specified file path
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
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
		log.Printf("Error reciving message: %v", err)
		return err
	}
	username := msg.GetSender()
	if !cs.isLoggedIn(username) {
		log.Printf("Unlogged in user: %s is not permit to chat!!!\n", username)
		return errors.New("Unlogged in user: " + username + " is not permit to chat!!!")
	}

	token := msg.GetMessage()
	log.Printf("User %s joined the chat room with token %s!\n", username, token)

	// TODO: validate token (later)

	cs.addClientStream(username, stream)

	/*
		other messages received from client will be used as messages in chat
	*/
	for {
		// Recieve messages from client
		msg, err = stream.Recv()
		if err != nil {
			s, _ := status.FromError(err)
			switch s.Code() {
			case codes.Canceled:
				log.Printf("Client disconnected: %s!\n", username)
				cs.removeClientStream(username)
				cs.loggedInAccount[username] = false
				return err
			default:
				log.Printf("Error reciving message: %v", err)
				cs.removeClientStream(username)
				return err
			}
		}

		/*
			Check for previous message likes.
			If previous message are not enough 2 likes, prevent them from sending the message
		*/
		log.Printf("Room chat request from %s: %s\n", msg.GetSender(), msg.GetMessage())

		if cs.messageLikes[username] < 2 {
			log.Printf("User %s has not enough likes to send new message to the room!\n", username)

			err := stream.Send(&gs.ChatMessage{
				Message: "Get more likes to send messages to room!!!",
				Sender:  "Server",
			})

			if err != nil {
				log.Printf("Error sending message to %s %v\n", username, err)
			}

			continue
		}

		// Broadcast message to all other users
		for recipient, recvStream := range cs.clientStream {
			if recipient != msg.GetSender() {
				err := recvStream.Send(&gs.ChatMessage{
					Message: msg.GetMessage(),
					Sender:  msg.GetSender(),
				})

				if err != nil {
					log.Printf("Error sending message to %s %v\n", recipient, err)
				}
			}
		}
		cs.messageLikes[username] = 0
	}
}

func (cs *ChatServer) LikeComment(ctx context.Context, command *gs.Command) (*gs.SentMessageStatus, error) {
	cmd := strings.Split(*command.AdditionalInfo, " ")
	sender := cmd[0]
	recipent := cmd[1]
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.messageLikes[recipent]++

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, sender)

	log.Printf("User %s just liked comment of %s\n", sender, recipent)

	return &gs.SentMessageStatus{
		Id:        id,
		Timestamp: timestamp,
		Status:    int32(codes.OK),
	}, nil
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

	if err != nil {
		log.Printf("Failed sending message from %s to %s: %s\n", msg.Sender, msg.Recipent, err.Error())
	} else {
		log.Printf("Message sent from %s to %s successfully\n", msg.Sender, msg.Recipent)
	}

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
		log.Printf("Failed to login as %s: Already login from another place!", in.Username)
		return &result, nil
	} else if in.Password != cs.serverPassword {
		msg := fmt.Sprintf("Failed to login as %s: Wrong password!", in.Username)
		result.Message = &msg
		log.Printf("Failed to login as %s: Wrong password!", in.Username)
		return &result, nil
	}

	msg := "Login successfully!!!!"
	result.Message = &msg
	log.Printf("%s was just logged in\n", in.Username)
	result.Status = int32(codes.OK)
	cs.loggedInAccount[in.Username] = true
	cs.messageLikes[in.Username] = 2

	return &result, nil
}

func (cs *ChatServer) Register(ctx context.Context, user *gs.User) (*gs.AuthenticationResult, error) {
	for _, registeredUser := range cs.registeredAccount.User {
		if registeredUser.Username == user.Username {

			msg := fmt.Sprintf("Username %s is already taken!", user.Username)
			authenticateMsg := gs.AuthenticationResult{
				Username: user.Username,
				Status:   int32(codes.AlreadyExists),
				Message:  &msg,
			}

			log.Printf("Username %s is already taken! Register failed\n", user.Username)
			return &authenticateMsg, errors.New("Username " + user.Username + " is already taken!")
		}
	}
	cs.registeredAccount.User = append(cs.registeredAccount.User, user)
	err := cs.writeUserInformation(cs.pathToUserCredentials)

	if err == nil {
		log.Printf("%s was just registered successfully\n", user.Username)
		return &gs.AuthenticationResult{Username: user.Username, Status: int32(codes.OK)}, nil
	} else {
		log.Printf("Register failed: %s\n", err.Error())
		return &gs.AuthenticationResult{Username: user.Username, Status: int32(codes.Internal)}, err
	}
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
	cs.messageLikes = make(map[string]int)
	cs.mu = sync.Mutex{}
	cs.serverPassword = serverPassword
	cs.pathToUserCredentials = pathToUserCredentials
	cs.loadUserInformation(cs.pathToUserCredentials)

	return &cs
}
