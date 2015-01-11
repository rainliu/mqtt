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
	GetReturnCodes() []byte
	SetReturnCodes([]byte)
}

////////////////////Implementation////////////////////////

type packet_suback struct {
	packet

	packetId    uint16
	returnCodes []byte
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
	remainingLength := uint32(2 + len(this.returnCodes))
	x, _ := this.EncodingRemainingLength(remainingLength)
	buffer.Write(x)

	//Variable Header
	buffer.WriteByte(byte(this.packetId >> 8))
	buffer.WriteByte(byte(this.packetId & 0xFF))

	//Payload
	buffer.Write(this.returnCodes)

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
	this.returnCodes = make([]byte, remainingLength-2)
	copy(this.returnCodes, buffer[consumedBytes:consumedBytes+remainingLength-2])
	for i := 0; i < int(remainingLength-2); i++ {
		if !(this.returnCodes[i] <= 0x02 || this.returnCodes[i] == 0x80) {
			return fmt.Errorf("Invalid %x Control Packet Return Code %02x\n", this.packetType, this.returnCodes[i])
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
func (this *packet_suback) GetReturnCodes() []byte {
	return this.returnCodes
}
func (this *packet_suback) SetReturnCodes(returnCodes []byte) {
	this.returnCodes = make([]byte, len(returnCodes))
	copy(this.returnCodes, returnCodes)
}
