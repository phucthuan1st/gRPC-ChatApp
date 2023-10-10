package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/phucthuan1st/gRPC-ChatRoom/grpcService"
	be "github.com/phucthuan1st/gRPC-ChatRoom/server/backend"
	"google.golang.org/grpc"
)

const (
	port           = 55555
	connectionType = "tcp"
	serverAddress  = "localhost"
)

func setupLogging(logFile *os.File) {

	// Create a multi-writer to write log messages to both the file and the terminal.
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Set the log output to the multi-writer.
	log.SetOutput(multiWriter)

	// Set log flags as desired (e.g., log date and time).
	log.SetFlags(log.Ldate | log.Ltime)
}

func main() {
	// Generate a log file name based on the current date and time.
	currentTime := time.Now()
	logFileName := fmt.Sprintf("server/log/app_%s.log", currentTime.Format("20060102_150405"))
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening log file: %v", err)
	}

	// Set up logging with the generated log file name.
	setupLogging(logFile)
	defer logFile.Close()

	// https://grpc.io/docs/languages/go/basics/#starting-the-server
	listener, err := net.Listen(connectionType, fmt.Sprintf("%s:%d", serverAddress, port))

	if err != nil {
		log.Fatalf("Cannot start the server: %s", err.Error())
		return
	}

	grpcServer := grpc.NewServer()
	backendServer := be.NewChatServer(os.Args[1])
	grpcService.RegisterChatRoomServer(grpcServer, backendServer)

	defer grpcServer.GracefulStop()

	// Start the gRPC server
	log.Printf("Starting gRPC server on localhost:%d...", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
