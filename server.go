package mqtt

import "net"

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	Forward(msg Message) error
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

func (this *serverSession) Forward(msg Message) error {
	//TODO: how to filter topic?
	_, err := this.conn.Write(msg.Packetize(0).Bytes())
	return err
}

func (this *serverSession) Process(pkt []byte) Event {
	switch this.state {
	case SESSION_STATE_CREATED:
	case SESSION_STATE_ACTIVE:
	case SESSION_STATE_TERMINATED:
	default:
	}

	return nil
}
