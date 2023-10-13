# gRPC Golang Chat Application

- Author: Nguyễn Phúc Thuần
- Org: FIT@HCMUS
- StudentID: 20120380

## Overview

This is a simple chat application built using gRPC (Google Remote Procedure Call) in Go with a graphical user interface (GUI) powered by rivo/tview. It allows users to connect and chat with each other in real-time.

The application demonstrates the use of gRPC for communication between a client and server while offering a user-friendly interface for a seamless chatting experience.

## Technologies Used

- **Go (Golang):** Go is a statically typed, compiled language known for its performance and simplicity. It serves as the foundation for this project.
- **gRPC:** gRPC is a high-performance RPC (Remote Procedure Call) framework developed by Google. It is used to establish efficient communication between the client and server.
- **tview:** It enables the creation of graphical user interfaces for applications, making it easier for users to interact with the chatroom. Powered by rivo terminal framework.

## Features

- Real-time chat with multiple users.
- gRPC-based communication for efficient and fast messaging.
- User-friendly graphical interface powered by tview.
- Simple and easy-to-use command-line interface for setting up and running the application.

## Prerequisites

Before you can run this application, make sure you have the following installed:

- **Go (Golang):** You can download and install Go from [here](https://golang.org/dl/).
- **gRPC for Go:** Follow the gRPC installation guide for Go [here](https://grpc.io/docs/languages/go/quickstart/).
- **tview:** Install tview by following the instructions in the [tview repository](https://github.com/rivo/tview).

But do not worry. I will include gRPC and tview in go module config (go.mod) and vendor, just follow the Usage for an easy use.

<br>

## Usage

### Install

1. Clone the repository:

```
git clone https://github.com/phucthuan1st/gRPC-ChatRoom.git
```

2. Change directory to the project folder:

```
cd gRPC-ChatRoom
```

3. Install dependencies (Optional if **vendor** folder included):

```
go get -u all
```

4. Start the server:

```
go run server/main.go
```
Flags:
```
-server                        : gRPC server (or callee) address, default: localhost
-port                            : server (or callee) port, default: 55555
-connectionType   : type of connection between stubs, default: tcp
-credDB                     : path to json contains user credentials and infomation, default: db/UserCredentials.json
-logDir                        : specify where should the server put the log file on
```

5. Start a client (multiple clients can be run in different terminal windows):

```
go run client/main.go
```
Flags:
```
-ipaddr                      : gRPC server (or callee) address, default: localhost
-port                           : server (or callee) port, default: 55555
-interval                    : application refresh/update interval, default 100*Millisecond
```

6. Follow the on-screen instructions to chat with other users using the tview GUI.

### Run without install

You can download the release version from [Releases here](https://github.com/phucthuan1st/gRPC-ChatRoom/releases/tag/beta) 

Else if you're already clone the repo, you can just run the server by navigate to bin folder, then run the program (in terminal)
```
./server [FLAG]
```

and the client app:
```
./client [FLAG]
```

If there is no binary pre-built file in the bin folder, or you just wanna build binary/executable file in you system, quickly run following command in the project folder (gRPC-ChatRoom)
```
go build -o bin/server server/main.go && go build -o bin/client client/main.go
```

## Protobuf
If you wanna update the protobuf, you can do it by simply run [gen_protobuf script](https://github.com/phucthuan1st/gRPC-ChatRoom/blob/master/gen_protobuf.sh)
<br>

## Contributing

Contributions are welcome! If you'd like to contribute to this project, please follow these guidelines:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes and test thoroughly.
4. Create a pull request with a clear description of your changes.

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/phucthuan1st/gRPC-ChatRoom/blob/master/LICENSE) file for details.

## Acknowledgments

- The gRPC team for providing a powerful and efficient communication framework.
- The Go community for their support and contributions.
- The rivo's tview developers for enabling graphical user interfaces in Go applications.

## Contact

If you have any questions or suggestions, please feel free to contact us at phucthuan.work@gmail.com

<br>

<strong>Happy chatting with gRPC and tview! Let's GO!</strong>
