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
	Acknowledge(pktack PacketAck) error
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
	serverPacketId uint16
	clientPacketId uint16
	PacketIds      map[uint32]uint16

	//Subscribe
	topics map[string]string
	qos    map[string]QOS

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
	this.clientPacketId = 1
	this.serverPacketId = 1
	this.PacketIds = make(map[uint32]uint16)
	this.keepAlive = 0
	this.keepAliveAccumulated = 0
	this.topics = make(map[string]string)
	this.qos = make(map[string]QOS)

	return this
}

func (this *serverSession) Forward(msg Message) error {
	if this.state == SESSION_STATE_CONNECT || this.state == SESSION_STATE_PUBLISH {
		if _, ok := this.topics[msg.GetTopic()]; ok {
			if _, err := this.conn.Write(msg.Packetize(this.serverPacketId).Bytes()); err != nil {
				log.Println(err.Error())
				return err
			}

			if msg.GetQos() == QOS_TWO || msg.GetQos() == QOS_ONE {
				this.state = SESSION_STATE_PUBLISH
				this.PacketIds[uint32(this.serverPacketId)] = this.serverPacketId
				if this.serverPacketId++; this.serverPacketId == 0 {
					this.serverPacketId++
				}
			}
		}
	}

	return nil
}

func (this *serverSession) Acknowledge(pktack PacketAck) error {
	switch this.state {
	case SESSION_STATE_CREATED:
		pkgconnack, ok := pktack.(PacketConnack)
		if !ok {
			return errors.New("Invalid PacketConnack in SESSION_STATE_CREATED\n")
		}
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			log.Println(err.Error())
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
		pkgsuback, ok := pktack.(PacketSuback)
		if !ok {
			return errors.New("Invalid PacketSuback in SESSION_STATE_CONNECT\n")
		}
		retCodes := pkgsuback.GetReturnCodes()
		if len(this.qosToBeAdded) != len(retCodes) {
			return errors.New("Invalid Return Codes Length in PacketSuback\n")
		}
		if _, err := this.conn.Write(pkgsuback.Bytes()); err != nil {
			log.Println(err.Error())
			return err
		} else {
			log.Println("SENT SUBACK")
		}
		for i := 0; i < len(retCodes); i++ {
			if retCodes[i] <= 0x02 {
				this.topics[this.topicsToBeAdded[i]] = this.topicsToBeAdded[i]
				this.qos[this.topicsToBeAdded[i]] = QOS(retCodes[i])
			}
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
				log.Println(err.Error())
			} else {
				log.Println("SENT PINGRESP")
			}
		case PACKET_DISCONNECT:
			return this.ProcessTerminate("DISCONNECT Packet Received\n")
		default:
			return this.ProcessTerminate(fmt.Sprintf("Unexpected %s Packet Received\n", PACKET_TYPE_STRINGS[pkt.GetType()]))
		}
	case SESSION_STATE_PUBLISH:
		switch pkt.GetType() {
		case PACKET_PUBREL:
			if pkt.(PacketPubrel).GetPacketId() != this.clientPacketId {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubRel PacketId %x Received\n", pkt.(PacketPubrec).GetPacketId()))
			} else {
				if delete(this.PacketIds, uint32(this.clientPacketId)<<16); len(this.PacketIds) == 0 {
					this.state = SESSION_STATE_CONNECT
				}
				pkgpubcomp := NewPacketAcks(PACKET_PUBCOMP)
				pkgpubcomp.SetPacketId(this.clientPacketId)
				if _, err := this.conn.Write(pkgpubcomp.Bytes()); err != nil {
					log.Println(err.Error())
				} else {
					log.Println("SENT PUBCOMP")
				}
			}
		case PACKET_PUBACK:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubAck PacketId %x Received\n", serverPacketId))
			} else {
				if delete(this.PacketIds, serverPacketId); len(this.PacketIds) == 0 {
					this.state = SESSION_STATE_CONNECT
				}
			}
		case PACKET_PUBREC:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubRec PacketId %x Received\n", serverPacketId))
			} else {
				pkgpubrel := NewPacketAcks(PACKET_PUBREL)
				pkgpubrel.SetPacketId(uint16(serverPacketId))
				if _, err := this.conn.Write(pkgpubrel.Bytes()); err != nil {
					log.Println(err.Error())
				} else {
					log.Println("SENT PUBREL")
				}
			}
		case PACKET_PUBCOMP:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubComp PacketId %x Received\n", serverPacketId))
			} else {
				if delete(this.PacketIds, serverPacketId); len(this.PacketIds) == 0 {
					this.state = SESSION_STATE_CONNECT
				}
			}
		default:
			return this.ProcessTerminate(fmt.Sprintf("Unexpected %s Packet Received\n", PACKET_TYPE_STRINGS[pkt.GetType()]))
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
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("SENT CONNACK")
		}

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
		this.clientPacketId = pktpub.GetPacketId()
		this.PacketIds[uint32(this.clientPacketId)<<16] = this.clientPacketId
		pkgpubrec := NewPacketAcks(PACKET_PUBREC)
		pkgpubrec.SetPacketId(this.clientPacketId)
		if _, err := this.conn.Write(pkgpubrec.Bytes()); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("SENT PUBREC")
		}
	} else if qos == QOS_ONE {
		pkgpuback := NewPacketAcks(PACKET_PUBACK)
		pkgpuback.SetPacketId(pktpub.GetPacketId())
		if _, err := this.conn.Write(pkgpuback.Bytes()); err != nil {
			log.Println(err.Error())
		} else {
			log.Println("SENT PUBACK")
		}
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
	for i := 0; i < len(topics); i++ {
		delete(this.topics, topics[i])
		delete(this.qos, topics[i])
	}

	pkgsuback := NewPacketAcks(PACKET_UNSUBACK)
	pkgsuback.SetPacketId(pktunsub.GetPacketId())
	if _, err := this.conn.Write(pkgsuback.Bytes()); err != nil {
		log.Println(err.Error())
	} else {
		log.Println("SENT UNSUBACK")
	}

	return newEventUnsubscribe(this, topics)
}

func (this *serverSession) ProcessTerminate(msg string) Event {
	this.state = SESSION_STATE_TERMINATED
	this.err = errors.New(msg)
	return newEventSessionTerminated(this, msg)
}
