package main

import (
	"crypto/tls"
	"mqtt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 4 {
		print("Usage: mqtt_server tcp 192.0.0.1 1883")
		return
	}

	var network string
	var address string
	var port int
	var tlsc *tls.Config
	var err error

	network = os.Args[1]
	address = os.Args[2]
	if port, err = strconv.Atoi(os.Args[3]); err != nil {
		print("Invalid port number")
		return
	}
	tlsc = nil

	stack := mqtt.GetStack()
	provider := stack.CreateProvider()

	transport := stack.CreateTransport(network, address, port, tlsc)
	provider.AddServerTransport(transport.(mqtt.ServerTransport))

	listener := newListener(provider)
	provider.AddListener(listener)

	stack.Run()
}
