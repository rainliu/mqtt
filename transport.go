package mqtt

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"time"
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
	GetTLSConfig() *tls.Config
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
	tlsc    *tls.Config

	//for server
	lner net.Listener
	ch   chan bool
}

func newTransport(network string, address string, port int, tlsc *tls.Config) *transport {
	this := &transport{}

	this.network = network
	this.address = address
	this.port = port
	this.tlsc = tlsc

	this.lner = nil
	this.ch = make(chan bool)

	return this
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

func (this *transport) GetTLSConfig() *tls.Config {
	return this.tlsc
}

//Client Transport
func (this *transport) Dial() (net.Conn, error) {
	var conn net.Conn
	var err error

	switch this.network {
	case TCP:
		conn, err = net.Dial("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)))
	case SSL:
		fallthrough
	case TCPS:
		fallthrough
	case TLS:
		conn, err = tls.Dial("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)), this.tlsc)
	}

	return conn, err
}

//Sever Transport
func (this *transport) Listen() error {
	var err error

	switch this.network {
	case TCP:
		this.lner, err = net.Listen("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)))
	case SSL:
		fallthrough
	case TCPS:
		fallthrough
	case TLS:
		this.lner, err = tls.Listen("tcp", net.JoinHostPort(this.address, strconv.Itoa(this.port)), this.tlsc)
	}

	return err
}

func (this *transport) Accept() (net.Conn, error) {
	if this.lner != nil {
		var conn net.Conn
		var err error

		switch this.network {
		case TCP:
			fallthrough
		case SSL:
			fallthrough
		case TCPS:
			fallthrough
		case TLS:
			conn, err = this.lner.Accept()
		}

		return conn, err
	} else {
		return nil, errors.New("Listen() must be called first or Listener is nil\n")
	}
}

func (this *transport) Close() {
	if this.lner != nil {
		close(this.ch)
		this.lner.Close()
	}
}

func (this *transport) SetDeadline(t time.Time) error {
	if tcpln, ok := this.lner.(*net.TCPListener); ok {
		return tcpln.SetDeadline(t)
	} else {
		return errors.New("Listener doesn't support SetDeadline\n")
	}
}
