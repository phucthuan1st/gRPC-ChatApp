package main

import (
	"github.com/phucthuan1st/gRPC-ChatRoom/client/app"
	"flag"
	"time"
)

var (
	port int = 55555
	ipaddr = "localhost"
	refreshInterval = time.Millisecond * 100
)

func main() {
	flag.StringVar(&ipaddr, "ipaddr", ipaddr, "server ip address")
	flag.IntVar(&port, "port", port, "connection port")
	flag.DurationVar(&refreshInterval, "interval", refreshInterval, "app refresh interval")

	flag.Parse()

	client := app.ClientApp{
		RefreshInterval: refreshInterval,
		Port: port,
		Ipaddr: ipaddr,
	}
	client.Start()
	defer client.Exit()
}
