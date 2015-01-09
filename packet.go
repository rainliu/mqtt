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

	EncodingRemainingLength(X uint32) ([]byte, error)
	DecodingRemainingLength([]byte) (uint32, uint32, error)

	EncodingUTF8(U string) []byte
	DecodingUTF8([]byte) (string, uint32, error)

	EncodingBinary(B []byte) []byte
	DecodingBinary([]byte) ([]byte, uint32, error)

	//Fixed Header
	GetType() PacketType
	SetType(PacketType)

	GetFlags() byte
	SetFlags(byte)
}

type PacketPingreq interface {
	Packet
}

type PacketPingresp interface {
	Packet
}

type PacketDisconnect interface {
	Packet
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
		//pkt = NewPacketConnect()
	case PACKET_CONNACK:
		//pkt = NewPacketConnack()
	case PACKET_PUBLISH:
		//pkt = NewPacketPublish()
	case PACKET_PUBACK:
		pkt = NewPacketAck(PACKET_PUBACK)
	case PACKET_PUBREC:
		pkt = NewPacketAck(PACKET_PUBREC)
	case PACKET_PUBREL:
		pkt = NewPacketAck(PACKET_PUBREL)
	case PACKET_PUBCOMP:
		pkt = NewPacketAck(PACKET_PUBCOMP)
	case PACKET_SUBSCRIBE:
		//pkt = NewPacketSubscribe()
	case PACKET_SUBACK:
		pkt = NewPacketSuback()
	case PACKET_UNSUBSCRIBE:
		//pkt = NewPacketUnsubscribe()
	case PACKET_UNSUBACK:
		pkt = NewPacketAck(PACKET_UNSUBACK)
	case PACKET_PINGREQ:
		pkt = NewPacket(PACKET_PINGREQ)
	case PACKET_PINGRESP:
		pkt = NewPacket(PACKET_PINGRESP)
	case PACKET_DISCONNECT:
		pkt = NewPacket(PACKET_DISCONNECT)
	default:
		return nil, fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}

	if pkt == nil {
		return nil, errors.New("Can't NewPacket")
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

func NewPacket(pt PacketType) *packet {
	if !(pt == PACKET_PINGREQ || pt == PACKET_PINGRESP || pt == PACKET_DISCONNECT) {
		return nil
	}

	this := packet{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = pt
	if pt == PACKET_PUBREL || pt == PACKET_SUBSCRIBE || pt == PACKET_UNSUBSCRIBE {
		this.packetFlag = 2
	} else {
		this.packetFlag = 0
	}

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
	if buffer == nil || len(buffer) != 2 {
		return errors.New("Invalid Control Packet Size")
	}

	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid Control Packet Flags %d\n", packetFlag)
	}
	if buffer[1] != 0 {
		return fmt.Errorf("Invalid Control Packet Remaining Length %d\n", buffer[1])
	}

	return nil
}

func (this *packet) Parse(buffer []byte) error {
	return this.IParser.IParse(buffer)
}

func (this *packet) EncodingRemainingLength(X uint32) ([]byte, error) {
	if X > 0xFFFFFF7F {
		return nil, errors.New("X value > 0xFFFFFF7F")
	}

	var buffer bytes.Buffer
	var encodedByte byte
	for X > 0 {
		encodedByte = byte(X % 128)
		X = X / 128
		if X > 0 {
			encodedByte = encodedByte | 128
		} else {
			buffer.WriteByte(encodedByte)
		}
	}

	return buffer.Bytes(), nil
}
func (this *packet) DecodingRemainingLength(buffer []byte) (uint32, uint32, error) {
	multipler := uint32(1)
	encodedByte := byte(128)

	value := uint32(0)
	i := 0
	for encodedByte&128 != 0 {
		if len(buffer) > i {
			encodedByte = buffer[i]
			i++
			value += uint32(encodedByte&127) * multipler
			multipler *= 128
			if multipler > 128*128*128 {
				return 0, 0, errors.New("Malformed Remaining Length")
			}
		} else {
			return 0, 0, errors.New("Malformed Remaining Length")
		}
	}

	return value, uint32(i), nil
}

func (this *packet) EncodingUTF8(U string) []byte {
	var buffer bytes.Buffer

	length := uint16(len(U))
	buffer.WriteByte(byte(length >> 8))
	buffer.WriteByte(byte(length & 0xFF))

	buffer.WriteString(U)

	return buffer.Bytes()
}
func (this *packet) DecodingUTF8(buffer []byte) (string, uint32, error) {
	if len(buffer) < 2 {
		return "", 0, errors.New("Malformed UTF8 encoded strings")
	}

	length := ((uint16(buffer[0])) << 8) | uint16(buffer[1])
	if len(buffer) < int(2+length) {
		return "", 0, errors.New("Malformed UTF8 encoded strings")
	}

	return string(buffer[2 : 2+length]), uint32(2 + length), nil
}

func (this *packet) EncodingBinary(B []byte) []byte {
	var buffer bytes.Buffer

	length := uint16(len(B))
	buffer.WriteByte(byte(length >> 8))
	buffer.WriteByte(byte(length & 0xFF))

	buffer.Write(B)

	return buffer.Bytes()
}
func (this *packet) DecodingBinary(buffer []byte) ([]byte, uint32, error) {
	if len(buffer) < 2 {
		return nil, 0, errors.New("Malformed UTF8 encoded strings")
	}

	length := ((uint16(buffer[0])) << 8) | uint16(buffer[1])
	if len(buffer) < int(2+length) {
		return nil, 0, errors.New("Malformed UTF8 encoded strings")
	}

	return buffer[2 : 2+length], uint32(2 + length), nil
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
func (this *packet) SetFlags(pf byte) {
	this.packetFlag = pf
}
