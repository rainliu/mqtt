package mqtt

import "net"

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	Forward(message Message)
}

////////////////////Implementation////////////////////////

type serverSession struct {
	session

	conn net.Conn
}

func newServerSession(conn net.Conn) *serverSession {
	this := &serverSession{}

	this.conn = conn

	return this
}

func (this *serverSession) Forward(message Message) {

}

func (this *serverSession) Run() {
}
