package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

////////////////////Interface//////////////////////////////

type SubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetReturnCode() []byte
	SetReturnCode([]byte)
}

////////////////////Implementation////////////////////////

type packet_suback struct {
	packet

	remainingLength uint32
	packetId        uint16
	returnCode      []byte
}

func NewPacketSuback() *packet_suback {
	this := packet_suback{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = PACKET_SUBACK
	this.packetFlag = 0

	return &this
}

func (this *packet_suback) IBytize() []byte {
	var buffer bytes.Buffer

	//Fixed Header
	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	this.remainingLength = uint32(2 + len(this.returnCode))
	x, _ := this.EncodingRemainingLength(this.remainingLength)
	buffer.Write(x)

	//Variable Header
	buffer.WriteByte(byte(this.packetId >> 8))
	buffer.WriteByte(byte(this.packetId & 0xFF))

	//Payload
	buffer.Write(this.returnCode)

	return buffer.Bytes()
}

func (this *packet_suback) IParse(buffer []byte) error {
	var err error
	var consumedBytes uint32

	if buffer == nil || len(buffer) < 4 {
		return errors.New("Invalid Control Packet Size")
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid Control Packet Flags %d\n", packetFlag)
	}
	if this.remainingLength, consumedBytes, err = this.DecodingRemainingLength(buffer[1:]); err != nil {
		return err
	}
	consumedBytes += 1
	if len(buffer)-int(consumedBytes) < int(this.remainingLength) {
		return errors.New("Invalid Control Packet Size")
	}

	//Variable Header
	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	consumedBytes += 2

	//Payload
	this.returnCode = make([]byte, this.remainingLength-2)
	copy(this.returnCode, buffer[consumedBytes:consumedBytes+this.remainingLength-2])

	return nil
}

//Variable Header
func (this *packet_suback) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_suback) SetPacketId(id uint16) {
	this.packetId = id
}

//Payload
func (this *packet_suback) GetReturnCode() []byte {
	return this.returnCode
}
func (this *packet_suback) SetReturnCode(returnCode []byte) {
	this.returnCode = make([]byte, len(returnCode))
	copy(this.returnCode, returnCode)
}
