package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketAcks interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PacketPuback interface {
	PacketAcks
}

type PacketPubrec interface {
	PacketAcks
}

type PacketPubrel interface {
	PacketAcks
}

type PacketPubcomp interface {
	PacketAcks
}

type PacketUnsuback interface {
	PacketAcks
}

////////////////////Implementation////////////////////////

type packet_acks struct {
	packet

	packetId uint16
}

func NewPacketAcks(pt PacketType) *packet_acks {
	if !(pt == PACKET_PUBACK || pt == PACKET_PUBREC || pt == PACKET_PUBREL || pt == PACKET_PUBCOMP || pt == PACKET_UNSUBACK) {
		return nil
	}

	this := packet_acks{}

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

func (this *packet_acks) IBytize() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(2)
	buffer.WriteByte(byte(this.packetId >> 8))
	buffer.WriteByte(byte(this.packetId & 0xFF))

	return buffer.Bytes()
}

func (this *packet_acks) IParse(buffer []byte) error {
	if buffer == nil || len(buffer) != 4 {
		return fmt.Errorf("Invalid %x Control Packet Size %x\n", this.packetType, len(buffer))
	}

	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %x Control Packet Type %x\n", this.packetType, packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %x Control Packet Flags %x\n", this.packetType, packetFlag)
	}
	if buffer[1] != 2 {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, buffer[1])
	}

	this.packetId = ((uint16(buffer[2])) << 8) | uint16(buffer[3])

	return nil
}

//Variable Header
func (this *packet_acks) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_acks) SetPacketId(id uint16) {
	this.packetId = id
}
