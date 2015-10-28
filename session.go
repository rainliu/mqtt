package mqtt


import(
	"errors" 
 	"fmt" 
	"net" 
)
////////////////////Interface//////////////////////////////

type SessionState byte

const (
	SESSION_STATE_CREATED SessionState = iota
	SESSION_STATE_CONNECTED
	SESSION_STATE_TERMINATED
)

type Session interface {
	GetRetransmitTimer() int
	SetRetransmitTimer(retransmitTimer int)

	GetState() SessionState
	Error() string
	Terminate(err error)

	GetAppData() interface{}
	SetAppData(interface{})
	
	
	Will() Message
	Forward(msg Message) error
	AcknowledgeConnect(pktconnack PacketConnack) error
	AcknowledgeSubscribe(pktsuback PacketSuback) error
}

////////////////////Implementation////////////////////////

type session struct {
	state           SessionState
	err             error
	quit            chan bool
	appData         interface{}
	retransmitTimer int
	
	conn net.Conn
	tracer Tracer

	//Connect
	keepAlive uint16
	clientId  string
	will      Message

	//Publish
	packetId  uint16
	PacketIds map[uint32]uint16

	//Subscribe
	topics map[string]string
	qos    map[string]QOS

	//private
	keepAliveAccumulated uint16
	topicsToBeAdded      []string
	qosToBeAdded         []QOS
}

func newSession(conn net.Conn, tracer Tracer) *session {
	this := &session{}

	this.conn = conn
	this.tracer = tracer
	this.err = nil
	this.state = SESSION_STATE_CREATED
	this.quit = make(chan bool)
	this.packetId = 1
	this.PacketIds = make(map[uint32]uint16)
	this.keepAlive = 0
	this.keepAliveAccumulated = 0
	this.topics = make(map[string]string)
	this.qos = make(map[string]QOS)
	this.will = nil

	return this
}

func (this *session) GetRetransmitTimer() int {
	return this.retransmitTimer
}

func (this *session) SetRetransmitTimer(retransmitTimer int) {
	this.retransmitTimer = retransmitTimer
}

func (this *session) GetState() SessionState {
	return this.state
}

func (this *session) Error() string {
	return this.err.Error()
}

func (this *session) Terminate(err error) {
	this.state = SESSION_STATE_TERMINATED
	this.err = err
	close(this.quit)
}

func (this *session) GetAppData() interface{} {
	return this.appData
}

func (this *session) SetAppData(appData interface{}) {
	this.appData = appData
}

func (this *session) Forward(msg Message) error {
	if this.state == SESSION_STATE_CONNECTED { //&& msg.GetClientId() != this.clientId {
		for _, sub := range this.topics {
			if this.Match(sub, msg.GetTopic()) {
				if _, err := this.conn.Write(msg.Packetize(this.packetId).Bytes()); err != nil {
					this.tracer.Println(err.Error())
					return err
				}

				if msg.GetQos() == QOS_TWO || msg.GetQos() == QOS_ONE {
					this.PacketIds[uint32(this.packetId)] = this.packetId
					if this.packetId++; this.packetId == 0 {
						this.packetId++
					}
				}

				break
			}
		}
	}

	return nil
}

func (this *session) AcknowledgeConnect(pktconnack PacketConnack) error {
	switch this.state {
	case SESSION_STATE_CREATED:
		if _, err := this.conn.Write(pktconnack.Bytes()); err != nil {
			this.tracer.Println(err.Error())
			return err
		} else {
			this.tracer.Println("SENT CONNACK")
		}
		if pktconnack.GetReturnCode() == CONNACK_RETURNCODE_ACCEPTED {
			this.state = SESSION_STATE_CONNECTED
		} else {
			this.state = SESSION_STATE_TERMINATED
			this.err = fmt.Errorf("Listener Refused Connection with Return Code %x\n", pktconnack.GetReturnCode())
		}
		return nil
	default:
		return errors.New("Invalid ServerSession State\n")
	}
}

func (this *session) AcknowledgeSubscribe(pktsuback PacketSuback) error {
	switch this.state {
	case SESSION_STATE_CONNECTED:
		retCodes := pktsuback.GetReturnCodes()
		if len(this.qosToBeAdded) != len(retCodes) {
			return errors.New("Invalid Return Codes Length in PacketSuback\n")
		}
		if _, err := this.conn.Write(pktsuback.Bytes()); err != nil {
			this.tracer.Println(err.Error())
			return err
		} else {
			this.tracer.Println("SENT SUBACK")
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

func (this *session) Process(buf []byte) Event {
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
			return this.ProcessTerminate("Invalid First CONNECT Packet Received\n", false)
		}
	case SESSION_STATE_CONNECTED:
		switch pkt.GetType() {
		case PACKET_CONNECT:
			return this.ProcessTerminate("Invalid Second or Multiple CONNECT Packets Received\n", false)
		case PACKET_PUBLISH:
			return this.ProcessPublish(pkt.(PacketPublish))
		case PACKET_SUBSCRIBE:
			return this.ProcessSubscribe(pkt.(PacketSubscribe))
		case PACKET_UNSUBSCRIBE:
			return this.ProcessUnsubscribe(pkt.(PacketUnsubscribe))
		case PACKET_PINGREQ:
			this.tracer.Println("PINGREQ Packet Received")
			pkgpingresp := NewPacket(PACKET_PINGRESP)
			if _, err := this.conn.Write(pkgpingresp.Bytes()); err != nil {
				this.tracer.Println(err.Error())
			} else {
				this.tracer.Println("SENT PINGRESP")
			}
		case PACKET_PUBREL:
			clientPacketId := uint32(pkt.(PacketPubrel).GetPacketId()) << 16
			if _, ok := this.PacketIds[clientPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubRel PacketId %x Received\n", clientPacketId>>16), false)
			} else {
				delete(this.PacketIds, clientPacketId)
				pkgpubcomp := NewPacketAcks(PACKET_PUBCOMP)
				pkgpubcomp.SetPacketId(uint16(clientPacketId >> 16))
				if _, err := this.conn.Write(pkgpubcomp.Bytes()); err != nil {
					this.tracer.Println(err.Error())
				} else {
					this.tracer.Println("SENT PUBCOMP")
				}
			}
		case PACKET_PUBACK:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubAck PacketId %x Received\n", serverPacketId), false)
			} else {
				delete(this.PacketIds, serverPacketId)
			}
		case PACKET_PUBREC:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubRec PacketId %x Received\n", serverPacketId), false)
			} else {
				pkgpubrel := NewPacketAcks(PACKET_PUBREL)
				pkgpubrel.SetPacketId(uint16(serverPacketId))
				if _, err := this.conn.Write(pkgpubrel.Bytes()); err != nil {
					this.tracer.Println(err.Error())
				} else {
					this.tracer.Println("SENT PUBREL")
				}
			}
		case PACKET_PUBCOMP:
			serverPacketId := uint32(pkt.(PacketPuback).GetPacketId())
			if _, ok := this.PacketIds[serverPacketId]; !ok {
				return this.ProcessTerminate(fmt.Sprintf("Invalid PubComp PacketId %x Received\n", serverPacketId), false)
			} else {
				delete(this.PacketIds, serverPacketId)
			}
		case PACKET_DISCONNECT:
			return this.ProcessTerminate("DISCONNECT Packet Received\n", true)
		default:
			return this.ProcessTerminate(fmt.Sprintf("Unexpected %s Packet Received\n", PACKET_TYPE_STRINGS[pkt.GetType()]), false)
		}
	case SESSION_STATE_TERMINATED:
	default:
	}

	return nil
}

func (this *session) ProcessConnect(pkgconn PacketConnect) Event {
	if pkgconn.GetProtocolLevel() != 0x04 {
		pkgconnack := NewPacketConnack()
		pkgconnack.SetSPFlag(false)
		pkgconnack.SetReturnCode(CONNACK_RETURNCODE_REFUSED_UNACCEPTABLE_PROTOCOL_VERSION)
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			this.tracer.Println(err.Error())
		} else {
			this.tracer.Println("SENT CONNACK")
		}

		this.state = SESSION_STATE_TERMINATED
		this.err = fmt.Errorf("Invalid %x Control Packet Protocol Level %x\n", pkgconn.GetType(), pkgconn.GetProtocolLevel())
		return newEventSessionTerminated(this, this.Error(), nil)
	} else if len(pkgconn.GetClientId()) == 0 && (pkgconn.GetConnectFlags()&CONNECT_FLAG_CLEAN_SESSION) == 0 {
		pkgconnack := NewPacketConnack()
		pkgconnack.SetSPFlag(false)
		pkgconnack.SetReturnCode(CONNACK_RETURNCODE_REFUSED_IDENTIFIER_REJECTED)
		if _, err := this.conn.Write(pkgconnack.Bytes()); err != nil {
			this.tracer.Println(err.Error())
		} else {
			this.tracer.Println("SENT CONNACK")
		}

		this.state = SESSION_STATE_TERMINATED
		this.err = fmt.Errorf("Invalid %x Control Packet Identifier Rejected\n", pkgconn.GetType())
		return newEventSessionTerminated(this, this.Error(), nil)
	} else {
		this.keepAlive = pkgconn.GetKeepAlive()
		this.clientId = pkgconn.GetClientId()

		connectFlags := pkgconn.GetConnectFlags()
		willTopic := pkgconn.GetWillTopic()
		willMessage := pkgconn.GetWillMessage()

		if (connectFlags & CONNECT_FLAG_WILL_FLAG) != 0 {
			var retain bool
			if (connectFlags & CONNECT_FLAG_WILL_RETAIN) != 0 {
				retain = true
			} else {
				retain = false
			}
			this.will = NewMessage(false,
				QOS((connectFlags&(CONNECT_FLAG_WILL_QOS_BIT3|CONNECT_FLAG_WILL_QOS_BIT4))>>3),
				retain,
				willTopic,
				willMessage)
			//this.will.SetClientId(this.clientId)
		} else {
			this.will = nil
		}

		return newEventConnect(this, pkgconn)
	}
}

func (this *session) ProcessPublish(pktpub PacketPublish) Event {
	//pktpub.GetMessage().SetClientId(this.clientId)
	qos := pktpub.GetMessage().GetQos()
	if qos == QOS_TWO {
		clientPacketId := pktpub.GetPacketId()
		this.PacketIds[uint32(clientPacketId)<<16] = clientPacketId
		pkgpubrec := NewPacketAcks(PACKET_PUBREC)
		pkgpubrec.SetPacketId(clientPacketId)
		if _, err := this.conn.Write(pkgpubrec.Bytes()); err != nil {
			this.tracer.Println(err.Error())
		} else {
			this.tracer.Println("SENT PUBREC")
		}
	} else if qos == QOS_ONE {
		pkgpuback := NewPacketAcks(PACKET_PUBACK)
		pkgpuback.SetPacketId(pktpub.GetPacketId())
		if _, err := this.conn.Write(pkgpuback.Bytes()); err != nil {
			this.tracer.Println(err.Error())
		} else {
			this.tracer.Println("SENT PUBACK")
		}
	}
	return newEventPublish(this, pktpub.GetMessage())
}

func (this *session) ProcessSubscribe(pktsub PacketSubscribe) Event {
	this.topicsToBeAdded = make([]string, len(pktsub.GetSubscribeTopics()))
	copy(this.topicsToBeAdded, pktsub.GetSubscribeTopics())

	this.qosToBeAdded = make([]QOS, len(pktsub.GetQoSs()))
	copy(this.qosToBeAdded, pktsub.GetQoSs())

	return newEventSubscribe(this, pktsub.GetPacketId(), pktsub.GetSubscribeTopics(), pktsub.GetQoSs())
}

func (this *session) ProcessUnsubscribe(pktunsub PacketUnsubscribe) Event {
	topics := pktunsub.GetUnsubscribeTopics()
	for i := 0; i < len(topics); i++ {
		delete(this.topics, topics[i])
		delete(this.qos, topics[i])
	}

	pkgsuback := NewPacketAcks(PACKET_UNSUBACK)
	pkgsuback.SetPacketId(pktunsub.GetPacketId())
	if _, err := this.conn.Write(pkgsuback.Bytes()); err != nil {
		this.tracer.Println(err.Error())
	} else {
		this.tracer.Println("SENT UNSUBACK")
	}

	return newEventUnsubscribe(this, topics)
}

func (this *session) ProcessTerminate(msg string, disconnected bool) Event {
	if disconnected {
		this.will = nil
	}
	this.state = SESSION_STATE_TERMINATED
	this.err = errors.New(msg)
	return newEventSessionTerminated(this, msg, this.will)
}

func (this *session) Will() Message {
	return this.will
}

// Does a topic match a subscription?
func (this *session) Match(sub, topic string) bool {
	var slen, tlen int
	var spos, tpos int
	multilevel_wildcard := false

	slen = len(sub)
	tlen = len(topic)

	if slen != 0 && tlen != 0 {
		if sub[0] == '$' && topic[0] != '$' || (topic[0] == '$' && sub[0] != '$') {
			return false
		}
	}

	spos = 0
	tpos = 0

	for spos < slen && tpos < tlen {
		if sub[spos] == topic[tpos] {
			spos++
			tpos++
			if spos == slen && tpos == tlen {
				return true
			} else if tpos == tlen && spos == slen-1 && sub[spos] == '+' {
				spos++
				return true
			}
		} else {
			if sub[spos] == '+' {
				spos++
				for tpos < tlen && topic[tpos] != '/' {
					tpos++
				}
				if tpos == tlen && spos == slen {
					return true
				}
			} else if sub[spos] == '#' {
				multilevel_wildcard = true
				if spos+1 != slen {
					return false
				} else {
					return true
				}
			} else {
				return false
			}
		}
		if tpos == tlen-1 {
			/* Check for e.g. foo matching foo/# */
			if spos == slen-3 && sub[spos+1] == '/' && sub[spos+2] == '#' {
				multilevel_wildcard = true
				return true
			}
		}
	}

	if multilevel_wildcard == false && (tpos < tlen || spos < slen) {
		return false
	}

	return true
}
