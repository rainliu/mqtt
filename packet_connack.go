package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketConnack interface {
	Packet

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() byte
	SetReturnCode(c byte)
}

////////////////////Implementation////////////////////////

type packet_connack struct {
	packet

	spFlag     byte
	returnCode byte
}

func NewPacketConnack() *packet_connack {
	this := packet_connack{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = PACKET_CONNACK
	this.packetFlag = 0

	return &this
}

func (this *packet_connack) IBytize() []byte {
	var buffer bytes.Buffer

	//Fixed Header
	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(2)

	//Variable Header
	buffer.WriteByte(byte(this.spFlag))
	buffer.WriteByte(this.returnCode)

	return buffer.Bytes()
}

func (this *packet_connack) IParse(buffer []byte) error {
	if buffer == nil || len(buffer) != 4 {
		return errors.New("Invalid Control Packet Size")
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid Control Packet Flags %d\n", packetFlag)
	}
	if buffer[1] != 2 {
		return errors.New("Invalid Control Packet Length")
	}

	//Variable Header
	if buffer[2]&0xFE != 0 {
		return fmt.Errorf("Invalid Control Packet Flags %d\n", buffer[2]&0xFE)
	}
	this.spFlag = buffer[2] & 0x01
	this.returnCode = buffer[3]

	return nil
}

//Variable Header
func (this *packet_connack) GetSPFlag() bool {
	return this.spFlag != 0
}
func (this *packet_connack) SetSPFlag(spFlag bool) {
	if spFlag {
		this.spFlag = 1
	} else {
		this.spFlag = 0
	}
}

func (this *packet_connack) GetReturnCode() byte {
	return this.returnCode
}
func (this *packet_connack) SetReturnCode(c byte) {
	this.returnCode = c
}
