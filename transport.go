package mqtt

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

	Connect()
	Disconnect()
}

type ServerTransport interface {
	Transport

	Listen()
	Accept()
}

////////////////////Implementation////////////////////////

type transport struct {
	network string
	address string
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
func (this *transport) Connect() {

}

func (this *transport) Disconnect() {

}

//Sever Transport
func (this *transport) Listen() {

}

func (this *transport) Accept() {

}
