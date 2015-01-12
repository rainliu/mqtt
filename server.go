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

func (this *serverSession) Process(buf []byte) Event {
	pkt, err := Packetize(buf)
	if err != nil && this.state != SESSION_STATE_TERMINATED {
		this.state = SESSION_STATE_TERMINATED
		return newEventIOException(this, this.conn.RemoteAddr())
	}

	switch this.state {
	case SESSION_STATE_CREATED:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			this.state = SESSION_STATE_CONNECT
			return newEventConnect(this, pkt.(PacketConnect))
		default:
			this.state = SESSION_STATE_TERMINATED
		}
	case SESSION_STATE_CONNECT:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			this.state = SESSION_STATE_TERMINATED
		case PACKET_PUBLISH:
			if pkt.(PacketPublish).GetMessage().GetQos() != QOS_ZERO {
				this.state = SESSION_STATE_PUBLISH
			}
			return newEventPublish(this, pkt.(PacketPublish).GetMessage())
		case PACKET_SUBSCRIBE:
		case PACKET_UNSUBSCRIBE:
		case PACKET_DISCONNECT:
			this.state = SESSION_STATE_TERMINATED
		default:
		}
	case SESSION_STATE_PUBLISH:

	case SESSION_STATE_TERMINATED:
	default:
	}

	return nil
}
