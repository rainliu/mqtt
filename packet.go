package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketType byte

const (
	PACKET_RESERVED_0 PacketType = iota
	PACKET_CONNECT
	PACKET_CONNACK
	PACKET_PUBLISH
	PACKET_PUBACK
	PACKET_PUBREC
	PACKET_PUBREL
	PACKET_PUBCOMP
	PACKET_SUBSCRIBE
	PACKET_SUBACK
	PACKET_UNSUBSCRIBE
	PACKET_UNSUBACK
	PACKET_PINGREQ
	PACKET_PINGRESP
	PACKET_DISCONNECT
	PACKET_RESERVED_15
)

type IBytizer interface {
	IBytize() []byte
}

type IParser interface {
	IParse([]byte) error
}

type Packet interface {
	IBytizer
	Bytes() []byte

	IParser
	Parse([]byte) error

	//Fixed Header
	GetType() PacketType
	SetType(pt PacketType)

	GetFlags() byte
	SetFlags(flags byte)
}

////////////////////Implementation////////////////////////

func Packetize(buffer []byte) (pkt Packet, err error) {
	defer func() {
		if r := recover(); r != nil {
			pkt = nil
			err = errors.New(r.(string))
		}
	}()

	packetType := PacketType((buffer[0] >> 4) & 0x0F)
	switch packetType {
	case PACKET_CONNECT:
		pkt = NewPacketConnect()
	case PACKET_CONNACK:
		//pkt = NewPacketConnack()
	case PACKET_PUBLISH:
		//pkt = NewPacketPublish()
	case PACKET_PUBACK:
		//pkt = NewPacketPuback()
	case PACKET_PUBREC:
		//pkt = NewPacketPubrec()
	case PACKET_PUBREL:
		//pkt = NewPacketPubrel()
	case PACKET_PUBCOMP:
		//pkt = NewPacketPubcomp()
	case PACKET_SUBSCRIBE:
		//pkt = NewPacketSubscribe()
	case PACKET_SUBACK:
		//pkt = NewPacketSuback()
	case PACKET_UNSUBSCRIBE:
		//pkt = NewPacketUnsubscribe()
	case PACKET_UNSUBACK:
		//pkt = NewPacketUnsuback()
	case PACKET_PINGREQ:
		//pkt = NewPacketPingreq()
	case PACKET_PINGRESP:
		//pkt = NewPacketPingresp()
	case PACKET_DISCONNECT:
		pkt = NewPacketDisconnect()
	default:
		return nil, fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}

	if err = pkt.Parse(buffer); err != nil {
		return nil, err
	} else {
		return pkt, nil
	}
}

type packet struct {
	IBytizer
	IParser

	packetType PacketType
	packetFlag byte
}

func NewPacket() *packet {
	this := packet{}
	this.IBytizer = &this
	this.IParser = &this
	return &this
}

func (this *packet) IBytize() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(0)

	return buffer.Bytes()
}

func (this *packet) Bytes() []byte {
	return this.IBytizer.IBytize()
}

func (this *packet) IParse(buffer []byte) error {
	this.packetType = PacketType((buffer[0] >> 4) & 0x0F)
	this.packetFlag = buffer[0] & 0x0F
	return nil
}

func (this *packet) Parse(buffer []byte) error {
	return this.IParser.IParse(buffer)
}

//Fixed Header
func (this *packet) GetType() PacketType {
	return this.packetType
}

func (this *packet) SetType(pt PacketType) {
	this.packetType = pt
}

func (this *packet) GetFlags() byte {
	return this.packetFlag
}

func (this *packet) SetFlags(flags byte) {
	this.packetFlag = flags
}

type ConnackPacket interface {
	Packet

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() byte
	SetReturnCode(c byte)
}

type PublishPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetMessage() Message
	SetMessage(Message)
}

type PubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubrecPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubrelPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubcompPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type SubscribePacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetSubscribeTopics() []string
	SetSubscribeTopics([]string)

	GetQos() []QOS
	SetQos([]QOS)
}

type SubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetReturnCode() []byte
	SetReturnCode([]byte)
}

type UnsubscribePacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetUnsubscribeTopics() []string
	SetUnsubscribeTopics([]string)
}

type UnsubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PingreqPacket interface {
	Packet
}

type PingrespPacket interface {
	Packet
}
