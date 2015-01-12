package mqtt

import (
	"fmt"
	"net"
)

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	SetUserName(s string)
	SetPassword(s []byte)

	Forward(msg Message) error
}

////////////////////Implementation////////////////////////

type serverSession struct {
	session

	conn net.Conn

	//Server side
	userName string
	password []byte

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
	this.packetId = 0

	return this
}

func (this *serverSession) SetUserName(s string) {
	this.userName = s
}

func (this *serverSession) SetPassword(s []byte) {
	this.password = s
}

func (this *serverSession) Forward(msg Message) error {
	//TODO: how to filter topic?
	_, err := this.conn.Write(msg.Packetize(this.packetId).Bytes())
	return err
}

func (this *serverSession) Process(buf []byte) Event {
	pkt, err := Packetize(buf)
	if err != nil {
		if this.state != SESSION_STATE_TERMINATED {
			this.state = SESSION_STATE_TERMINATED
			this.err = err
			return newEventIOException(this, this.conn.RemoteAddr())
		} else {
			return nil
		}
	}

	switch this.state {
	case SESSION_STATE_CREATED:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			return this.ProcessConnect(pkt.(PacketConnect))
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

func (this *serverSession) ProcessConnect(pkgconn PacketConnect) Event {
	if pkgconn.GetProtocolLevel() != 0x04 {
		pkgconnack := NewPacketConnack()
		pkgconnack.SetSPFlag(false)
		pkgconnack.SetReturnCode(CONNACK_RETURNCODE_REFUSED_UNACCEPTABLE_PROTOCOL_VERSION)
		this.conn.Write(pkgconnack.Bytes())

		this.state = SESSION_STATE_TERMINATED
		this.err = fmt.Errorf("Invalid %x Control Packet Protocol Level %x\n", pkgconn.GetType(), pkgconn.GetProtocolLevel())
		return newEventSession(EVENT_SESSION_TERMINATED, this, this.Error())
	} else {
		this.connectFlags = pkgconn.GetConnectFlags()
		this.keepAlive = pkgconn.GetKeepAlive()
		this.clientId = pkgconn.GetClientId()
		this.willTopic = pkgconn.GetWillTopic()
		this.willMessage = pkgconn.GetWillMessage()
		this.state = SESSION_STATE_CONNECT
		return newEventConnect(this, pkgconn)
	}
}
