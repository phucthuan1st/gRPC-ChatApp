# gRPC Chat Room 

- Author: Nguyen Phuc Thuan
- Org: FIT@HCMUS
- CredID: 20120380

## Overview

This is a simple chat application built using gRPC (Google Remote Procedure Call) in Go with a graphical user interface (GUI) powered by go-gtk. It allows users to connect and chat with each other in real-time. The application demonstrates the use of gRPC for communication between a client and server while offering a user-friendly interface for a seamless chatting experience.

## Technologies Used

- **Go (Golang):** Go is a statically typed, compiled language known for its performance and simplicity. It serves as the foundation for this project.
- **gRPC:** gRPC is a high-performance RPC (Remote Procedure Call) framework developed by Google. It is used to establish efficient communication between the client and server.
- **go-gtk:** go-gtk is a Go binding for GTK, a popular GUI toolkit. It enables the creation of graphical user interfaces for applications, making it easier for users to interact with the chatroom.

## Features

- Real-time chat with multiple users.
- gRPC-based communication for efficient and fast messaging.
- User-friendly graphical interface powered by go-gtk.
- Simple and easy-to-use command-line interface for setting up and running the application.

## Prerequisites

Before you can run this application, make sure you have the following installed:

- **Go (Golang):** You can download and install Go from [here](https://golang.org/dl/).
- **gRPC for Go:** Follow the gRPC installation guide for Go [here](https://grpc.io/docs/languages/go/quickstart/).
- **go-gtk:** Install go-gtk by following the instructions in the [go-gtk repository](https://github.com/mattn/go-gtk).

<br>

## Usage

1. Clone the repository:
```
git clone https://github.com/phucthuan1st/gRPC-ChatRoom.git
```
2. Change directory to the project folder:
```
cd gRPC-ChatRoom
```
3. Install dependencies:
```
go get ./...
```
4. Start the server:
```
go run server/main.go
```
5. Start a client (multiple clients can be run in different terminal windows):
```
go run client/main.go
```
6. Follow the on-screen instructions to chat with other users using the go-gtk GUI.

<br> 

## Contributing

Contributions are welcome! If you'd like to contribute to this project, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and test thoroughly.
4. Create a pull request with a clear description of your changes.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- The gRPC team for providing a powerful and efficient communication framework.
- The Go community for their support and contributions.
- The go-gtk developers for enabling graphical user interfaces in Go applications.

## Contact

If you have any questions or suggestions, please feel free to contact us at phucthuan.work@gmail.com

<br>

Happy chatting with gRPC and go-gtk!
