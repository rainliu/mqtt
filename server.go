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

	//Connect
	connectFlags byte
	keepAlive    uint16
	clientId     string
	willTopic    string
	willMessage  string

	//Subscribe
	topics []string
	qos    []QOS

	//others
	packetId uint16
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
	_, err := this.conn.Write(msg.Packetize(this.packetId).Bytes())
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
