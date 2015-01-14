package mqtt

import (
	"errors"
	"fmt"
	"log"
	"net"
)

////////////////////Interface//////////////////////////////

type ServerSession interface {
	Session

	Forward(msg Message) error
	Acknowledge(pkt PacketAcks) error
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

	//private
	keepAliveAccumulated uint16
	topicsToBeAdded      []string
	qosToBeAdded         []QOS
}

func newServerSession(conn net.Conn) *serverSession {
	this := &serverSession{}

	this.conn = conn
	this.err = nil
	this.state = SESSION_STATE_CREATED
	this.ch = make(chan bool)
	//this.timeout = make(chan bool)
	this.packetId = 0
	this.keepAlive = 0
	this.keepAliveAccumulated = 0

	return this
}

func (this *serverSession) Forward(msg Message) error {
	if this.state == SESSION_STATE_CONNECT {
		//TODO: how to filter topic?
		_, err := this.conn.Write(msg.Packetize(this.packetId).Bytes())
		return err
	} else {
		return errors.New("Invalid ServerSession State\n")
	}
}

func (this *serverSession) Acknowledge(pkt PacketAcks) error {
	switch this.state {
	case SESSION_STATE_CREATED:
		pkgconnack, ok := pkt.(PacketConnack)
		if !ok {
			return errors.New("Invalid PacketConnack in SESSION_STATE_CREATED\n")
		}
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			return err
		} else {
			log.Println("SENT CONNACK")
		}
		if pkgconnack.GetReturnCode() == CONNACK_RETURNCODE_ACCEPTED {
			this.state = SESSION_STATE_CONNECT
		} else {
			this.state = SESSION_STATE_TERMINATED
			this.err = fmt.Errorf("Listener Refused Connection with Return Code %x\n", pkgconnack.GetReturnCode())
		}
		return nil
	case SESSION_STATE_CONNECT:
		pkgsuback, ok := pkt.(PacketSuback)
		if !ok {
			return errors.New("Invalid PacketSuback in SESSION_STATE_CONNECT\n")
		}
		if retCodes := pkgsuback.GetReturnCodes(); len(this.qosToBeAdded) != len(retCodes) {
			return errors.New("Invalid Return Codes Length in PacketSuback\n")
		} else {
			for i := 0; i < len(retCodes); i++ {
				if retCodes[i] <= 0x02 {
					this.topics = append(this.topics, this.topicsToBeAdded[i])
					this.qos = append(this.qos, QOS(retCodes[i]))
				}
			}
		}
		if _, err := this.conn.Write(pkgsuback.Bytes()); err != nil {
			return err
		} else {
			log.Println("SENT SUBACK")
		}
		return nil
	default:
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
			return this.ProcessTerminate("Invalid First CONNECT Packet Received\n")
		}
	case SESSION_STATE_CONNECT:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			return this.ProcessTerminate("Invalid Second or Multiple CONNECT Packets Received\n")
		case PACKET_PUBLISH:
			return this.ProcessPublish(pkt.(PacketPublish))
		case PACKET_SUBSCRIBE:
			return this.ProcessSubscribe(pkt.(PacketSubscribe))
		case PACKET_UNSUBSCRIBE:
			return this.ProcessUnsubscribe(pkt.(PacketUnsubscribe))
		case PACKET_PINGREQ:
			log.Println("PINGREQ Packet Received")
			pkgpingresp := NewPacket(PACKET_PINGRESP)
			if _, err := this.conn.Write(pkgpingresp.Bytes()); err != nil {
				log.Printf(err.Error())
			} else {
				log.Println("SENT PINGRESP")
			}
		case PACKET_DISCONNECT:
			return this.ProcessTerminate("DISCONNECT Packet Received\n")
		default:
			return this.ProcessTerminate(fmt.Sprintf("Unexpected %x Packet Received\n", pkt.GetType()))
		}
	case SESSION_STATE_PUBLISH:
		switch pkt.GetType() {
		case PACKET_PUBREL:
			if pkt.(PacketPubrel).GetPacketId() != this.pubPacketId {
				return this.ProcessTerminate("Invalid PubRel PacketId Received\n")
			} else {
				this.state = SESSION_STATE_CONNECT
				pkgpubcomp := NewPacketAcks(PACKET_PUBCOMP)
				pkgpubcomp.SetPacketId(this.pubPacketId)
				this.conn.Write(pkgpubcomp.Bytes())
			}
		default:
			return this.ProcessTerminate(fmt.Sprintf("Unexpected %x Packet Received\n", pkt.GetType()))
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
		return newEventSessionTerminated(this, this.Error())
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

func (this *serverSession) ProcessSubscribe(pktsub PacketSubscribe) Event {
	this.topicsToBeAdded = make([]string, len(pktsub.GetSubscribeTopics()))
	copy(this.topicsToBeAdded, pktsub.GetSubscribeTopics())

	this.qosToBeAdded = make([]QOS, len(pktsub.GetQoSs()))
	copy(this.qosToBeAdded, pktsub.GetQoSs())

	return newEventSubscribe(this, pktsub.GetPacketId(), pktsub.GetSubscribeTopics(), pktsub.GetQoSs())
}

func (this *serverSession) ProcessUnsubscribe(pktunsub PacketUnsubscribe) Event {
	topics := pktunsub.GetUnsubscribeTopics()

	pkgsuback := NewPacketAcks(PACKET_UNSUBACK)
	pkgsuback.SetPacketId(pktunsub.GetPacketId())
	this.conn.Write(pkgsuback.Bytes())

	return newEventUnsubscribe(this, topics)
}

func (this *serverSession) ProcessTerminate(msg string) Event {
	this.state = SESSION_STATE_TERMINATED
	this.err = errors.New(msg)
	return newEventSessionTerminated(this, msg)
}
