package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
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
	"google.golang.org/grpc/status"
)

var hasher = sha1.New()

type MessageLikes struct {
	whoLike map[string]bool
	nLike   int
}

type ChatServer struct {
	registeredAccount     gs.UserList
	loggedInAccount       map[string]bool
	clientStream          map[string]gs.ChatRoom_ChatServer
	messageLikes          map[string]MessageLikes
	mu                    sync.Mutex
	pathToUserCredentials string
	gs.UnimplementedChatRoomServer
}

// ---------------------------------------------------------//
// ------------------ HELPER -------------------------------//

// get the client stream for the given username
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

// Check if a peer is logged in and online
func (cs *ChatServer) isLoggedIn(username string) bool {
	online, ok := cs.loggedInAccount[username]
	return ok && online
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

// write the user information to a json file
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

// add the client stream to the connected client stream map
func (cs *ChatServer) addClientStream(username string, stream gs.ChatRoom_ChatServer) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.clientStream[username] = stream
}

// delete a client stream from the map when cleint is no longer online
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
	*/
	msg, err := stream.Recv()
	if err != nil {
		log.Printf("Error reciving message: %v", err)
		return err
	}

	username := msg.GetSender()
	log.Printf("User %s request to join the chat room!\n", username)

	// check if the user is already connected
	if !cs.isLoggedIn(username) {
		log.Printf("Unlogged in user: %s is not permit to chat!!!\n", username)
		return errors.New("Unlogged in user: " + username + " is not permit to chat!!!")
	}

	// welcome user to join the chat room
	stream.Send(&gs.ChatMessage{
		Message: fmt.Sprintf("Welcome to the chat room %s !", username),
		Sender:  "Server",
	})

	log.Printf("User %s is allowed to join the chat room!\n", username)
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
		log.Printf("Room chat request from %s: %s\n", username, msg.GetMessage())

		if cs.messageLikes[username].nLike < 2 {
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
		cs.mu.Lock()
		cs.broadcast(msg)
		cs.mu.Unlock()
	}
}

func (cs *ChatServer) broadcast(msg *gs.ChatMessage) {
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

	if msg.GetSender() != "Server" {
		cs.messageLikes[msg.GetSender()] = MessageLikes{
			nLike:   0,
			whoLike: make(map[string]bool),
		}
	}
}

// handle like command from client
func (cs *ChatServer) LikeMessage(ctx context.Context, command *gs.UserRequest) (*gs.SentMessageStatus, error) {
	sender := command.GetSender()
	recipent := command.GetTarget()

	cs.mu.Lock()
	defer cs.mu.Unlock()
	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, sender)
	var status codes.Code
	var err error

	messageLike, _ := cs.messageLikes[recipent]
	if messageLike.whoLike[sender] == true {
		log.Printf("User %s already liked message of %s and cannot like again\n", sender, recipent)

		cs.broadcast(&gs.ChatMessage{
			Message: fmt.Sprintf("%s requested to like Message of %s but rejected!!", sender, recipent),
			Sender:  "Server",
		})

		status = codes.AlreadyExists
		err = errors.New("Duplicated like request")
	} else {
		messageLike.nLike++
		messageLike.whoLike[sender] = true
		cs.messageLikes[recipent] = messageLike

		log.Printf("User %s just liked Message of %s\n", sender, recipent)
		cs.broadcast(&gs.ChatMessage{
			Message: fmt.Sprintf("%s just liked Message of %s", sender, recipent),
			Sender:  "Server",
		})

		status = codes.OK
		err = nil
	}

	return &gs.SentMessageStatus{
		Id:        id,
		Timestamp: timestamp,
		Status:    int32(status),
	}, err
}

// handle private message from client to client
func (cs *ChatServer) SendPrivateMessage(ctx context.Context, msg *gs.PrivateChatMessage) (*gs.SentMessageStatus, error) {

	log.Printf("%s sent a message to %s: %s\n", msg.Sender, msg.Recipent, msg.Message)

	timestamp := time.Now().Unix()
	id := fmt.Sprintf("%d-%s", timestamp, msg.Sender)
	status := 1

	stream, ok := cs.clientStream[msg.Recipent]
	if !ok {
		return nil, errors.New(fmt.Sprintf("User %s not found!", msg.Recipent))
	}

	var private int32 = 1

	err := stream.Send(&gs.ChatMessage{
		Sender:  msg.GetSender(),
		Message: msg.GetMessage(),
		Private: &private,
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
	hasher.Write([]byte(in.GetPassword()))

	log.Printf("Login request from %s: %s\n", in.Username, base64.StdEncoding.EncodeToString(hasher.Sum(nil)))

	result := gs.AuthenticationResult{}
	result.Username = in.Username

	// check for user that is already logged in or not
	if cs.isLoggedIn(in.Username) {
		msg := fmt.Sprintf("Failed to login as %s: Already login from another place!", in.Username)
		result.Message = &msg
		result.Status = int32(codes.AlreadyExists)

		log.Println(msg)
		return &result, errors.New(msg)
	}

	// allow user to login with correct username and password
	for _, user := range cs.registeredAccount.User {
		if user.Username != in.Username {
			continue
		}

		if in.Password != user.Password {
			msg := fmt.Sprintf("Failed to login as %s: Wrong password!", in.Username)
			result.Message = &msg
			result.Status = int32(codes.Unauthenticated)

			log.Println(msg)
			return &result, errors.New(msg)
		} else {
			// handle successful login
			cs.loggedInAccount[in.Username] = true
			cs.messageLikes[in.Username] = MessageLikes{
				nLike:   2,
				whoLike: make(map[string]bool),
			}

			msg := fmt.Sprintf("User %s has logged in successfully!", in.Username)
			result.Message = &msg
			result.Status = int32(codes.OK)

			log.Println(msg)
			cs.broadcast(&gs.ChatMessage{
				Message: msg,
				Sender:  "Server",
			})
			return &result, nil
		}
	}

	// other cases
	msg := fmt.Sprintf("Failed to login as %s: User not found!", in.Username)
	result.Status = int32(codes.Unauthenticated)
	return &result, errors.New(msg)
}

// handle register new account command from client
func (cs *ChatServer) Register(ctx context.Context, user *gs.User) (*gs.AuthenticationResult, error) {
	log.Printf("Register request from %s\n", user.Username)

	result := gs.AuthenticationResult{}
	result.Username = user.Username

	// Handle blank username or password case
	if user.Username == "" || user.Password == "" {
		msg := "Blank username or password is not allowed!"
		log.Println(msg)
		return nil, errors.New(msg)
	}

	// handle duplicate username case
	for _, registeredUser := range cs.registeredAccount.User {
		if registeredUser.Username == user.Username {

			msg := fmt.Sprintf("Username %s is already taken! Register failed", user.Username)
			authenticateMsg := gs.AuthenticationResult{
				Username: user.Username,
				Status:   int32(codes.AlreadyExists),
				Message:  &msg,
			}

			log.Println(msg)
			return &authenticateMsg, errors.New(msg)
		}
	}

	// validated case
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

// retrieve information about a user from connected users list
func (cs *ChatServer) GetPeerInfomations(ctx context.Context, cmd *gs.UserRequest) (*gs.PublicUserInfo, error) {
	sender := cmd.GetSender()
	target := cmd.GetTarget()

	log.Printf("%s requested information about %s\n", sender, target)

	for _, user := range cs.registeredAccount.User {
		if user.Username == target {
			result := &gs.PublicUserInfo{
				Username:  user.Username,
				FullName:  user.FullName,
				Address:   user.GetAddress(),
				Birthdate: user.Birthdate,
				Email:     user.Email,
			}

			return result, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("User %s not found or offline!", target))
}

func (cs *ChatServer) GetConnectedPeers(ctx context.Context, request *gs.UserRequest) (*gs.PublicUserInfoList, error) {
	sender := request.GetSender()
	log.Printf("%s requested connected users list\n", sender)

	result := &gs.PublicUserInfoList{}
	for _, user := range cs.registeredAccount.User {
		if user.Username == sender {
			continue
		}

		result.Username = append(result.Username, user.Username)
		if cs.isConnected(user.Username) {
			result.Status = append(result.Status, "Online")
		} else {
			result.Status = append(result.Status, "Offline")
		}
	}

	return result, nil
}

// ---------------------------------------------------------//

func GenerateSecureToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func NewChatServer(pathToUserCredentials string) *ChatServer {
	cs := ChatServer{}
	cs.clientStream = make(map[string]gs.ChatRoom_ChatServer)
	cs.loggedInAccount = make(map[string]bool)
	cs.messageLikes = make(map[string]MessageLikes)
	cs.mu = sync.Mutex{}
	cs.pathToUserCredentials = pathToUserCredentials
	cs.loadUserInformation(cs.pathToUserCredentials)

	return &cs
}
