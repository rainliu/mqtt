package mqtt

import "net"

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	Forward(m Message)
}

////////////////////Implementation////////////////////////

type serverSession struct {
	session

	conn net.Conn
}

func newServerSession(conn net.Conn) *serverSession {
	this := &serverSession{}

	this.conn = conn
	this.err = nil
	this.state = SESSION_STATE_CREATED
	this.ch = make(chan bool)

	return this
}

func (this *serverSession) Forward(m Message) {
	if _, err := this.conn.Write(m.Packetize()); err != nil {
		this.Terminate(err)
	}
}

func (this *serverSession) Process(p []byte) {

}
