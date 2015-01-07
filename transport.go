package mqtt

import (
	"errors"
	"net"
	"strconv"
)

////////////////////Interface//////////////////////////////

const (
	TCP  = "tcp"
	WS   = "ws"
	TLS  = "tls"
	SSL  = "ssl"
	TCPS = "tcps"

	PORT_1883 = 1883 //Non-TLS
	PORT_8883 = 8883 //TLS
)

type Transport interface {
	GetNetwork() string //"tcp", "tls", or "ws"...
	GetAddress() string
	GetPort() int
}

type ClientTransport interface {
	Transport

	Dial() (net.Conn, error)
}

type ServerTransport interface {
	Transport

	Listen() error
	Accept() (net.Conn, error)
	Close()
}

////////////////////Implementation////////////////////////

type transport struct {
	network string
	address string //for server, it is laddr; for client, it is raddr
	port    int
}

func newTransport(network string, address string, port int) *transport {
	return &transport{network: network, address: address, port: port}
}

func (this *transport) GetNetwork() string {
	return this.network
}

func (this *transport) GetAddress() string {
	return this.address
}

func (this *transport) GetPort() int {
	return this.port
}

//Client Transport
type clientTransport struct {
	transport
}

func (this *clientTransport) Dial() (net.Conn, error) {
	var conn net.Conn
	var err error

	switch this.network {
	case TCP:
		conn, err = net.Dial("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)))
	}

	return conn, err
}

//Sever Transport
type serverTransport struct {
	transport

	lner net.Listener
}

func (this *serverTransport) Listen() error {
	var err error

	switch this.network {
	case TCP:
		this.lner, err = net.Listen("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)))
	}

	return err
}

func (this *serverTransport) Accept() (net.Conn, error) {
	if this.lner != nil {
		var conn net.Conn
		var err error

		switch this.network {
		case TCP:
			conn, err = this.lner.Accept()
		}

		return conn, err
	} else {
		return nil, errors.New("Listen() must be called first or Listener is nil\n")
	}
}

func (this *serverTransport) Close() {
	if this.lner != nil {
		this.lner.Close()
	}
}
