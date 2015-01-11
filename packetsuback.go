package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketSuback interface {
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

	packetId   uint16
	returnCode []byte
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
	remainingLength := uint32(2 + len(this.returnCode))
	x, _ := this.EncodingRemainingLength(remainingLength)
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
	var remainingLength, consumedBytes uint32

	if buffer == nil || len(buffer) < 4 {
		return fmt.Errorf("Invalid %x Control Packet Size %x\n", this.packetType, len(buffer))
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %x Control Packet Type %x\n", this.packetType, packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %x Control Packet Flags %x\n", this.packetType, packetFlag)
	}
	if remainingLength, consumedBytes, err = this.DecodingRemainingLength(buffer[1:]); err != nil {
		return err
	}
	consumedBytes += 1
	if len(buffer)-int(consumedBytes) < int(remainingLength) {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, remainingLength)
	}

	//Variable Header
	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	consumedBytes += 2

	//Payload
	this.returnCode = make([]byte, remainingLength-2)
	copy(this.returnCode, buffer[consumedBytes:consumedBytes+remainingLength-2])
	for i := 0; i < int(remainingLength-2); i++ {
		if !(this.returnCode[i] <= 0x02 || this.returnCode[i] == 0x80) {
			return fmt.Errorf("Invalid %x Control Packet Return Code %02x\n", this.packetType, this.returnCode[i])
		}
	}

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
