package main

import (
	"github.com/phucthuan1st/gRPC-ChatRoom/client/app"
)

func main() {
	client := app.ClientApp{}
	client.Start()
	defer client.Exit()
}
