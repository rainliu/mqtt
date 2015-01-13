package mqtt

import (
	"errors"
	"fmt"
	"net"
)

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	Forward(msg Message) error
	Respond(spFlag bool, retCode CONNACK_RETURNCODE) error
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

	//Publish
	pubPacketId uint16

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

func (this *serverSession) Forward(msg Message) error {
	//TODO: how to filter topic?
	_, err := this.conn.Write(msg.Packetize(this.packetId).Bytes())
	return err
}

func (this *serverSession) Respond(spFlag bool, retCode CONNACK_RETURNCODE) error {
	if this.state == SESSION_STATE_CREATED {
		pkgconnack := NewPacketConnack()
		pkgconnack.SetSPFlag(spFlag)
		pkgconnack.SetReturnCode(retCode)
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			return err
		}
		if retCode == CONNACK_RETURNCODE_ACCEPTED {
			this.state = SESSION_STATE_CONNECT
		} else {
			this.state = SESSION_STATE_TERMINATED
			this.err = fmt.Errorf("Listener Refused Connection with Return Code %x\n", retCode)
		}
		return nil
	} else {
		return errors.New("Invalid ServerSession State\n")
	}
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
			this.err = errors.New("Invalid First CONNECT Packet Received\n")
		}
	case SESSION_STATE_CONNECT:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			this.state = SESSION_STATE_TERMINATED
			this.err = errors.New("Invalid Second or Multiple CONNECT Packets Received\n")
		case PACKET_PUBLISH:
			return this.ProcessPublish(pkt.(PacketPublish))
		case PACKET_SUBSCRIBE:
		case PACKET_UNSUBSCRIBE:
		case PACKET_DISCONNECT:
			this.state = SESSION_STATE_TERMINATED
		default:
		}
	case SESSION_STATE_PUBLISH:
		switch pkt.GetType() {
		case PACKET_PUBREL:
			pktpubrel := pkt.(PacketPubrel)
			if pktpubrel.GetPacketId() != this.pubPacketId {
				this.state = SESSION_STATE_TERMINATED
				this.err = errors.New("Invalid PubRel PacketId Received\n")
			} else {
				this.state = SESSION_STATE_CONNECT
				pkgpubcomp := NewPacketAcks(PACKET_PUBCOMP)
				pkgpubcomp.SetPacketId(this.pubPacketId)
				this.conn.Write(pkgpubcomp.Bytes())
			}
		default:
		}
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
		return newEventConnect(this, pkgconn)
	}
}

func (this *serverSession) ProcessPublish(pktpub PacketPublish) Event {
	qos := pktpub.GetMessage().GetQos()
	if qos == QOS_TWO {
		this.state = SESSION_STATE_PUBLISH
		this.pubPacketId = pktpub.GetPacketId()
		pkgpuback := NewPacketAcks(PACKET_PUBREC)
		pkgpuback.SetPacketId(this.pubPacketId)
		this.conn.Write(pkgpuback.Bytes())
	} else if qos == QOS_ONE {
		pkgpuback := NewPacketAcks(PACKET_PUBACK)
		pkgpuback.SetPacketId(pktpub.GetPacketId())
		this.conn.Write(pkgpuback.Bytes())
	}
	return newEventPublish(this, pktpub.GetMessage())
}
