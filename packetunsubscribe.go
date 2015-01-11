package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketUnsubscribe interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetUnsubscribeTopics() []string
	SetUnsubscribeTopics([]string)
}

////////////////////Implementation////////////////////////

type packet_unsubscribe struct {
	packet

	packetId uint16
	topics   []string
}

func NewPacketUnsubscribe() *packet_unsubscribe {
	this := packet_unsubscribe{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = PACKET_UNSUBSCRIBE
	this.packetFlag = 2

	return &this
}

func (this *packet_unsubscribe) IBytize() []byte {
	var buffer bytes.Buffer
	var buffer2 bytes.Buffer

	//1st Pass

	//Variable Header
	buffer2.WriteByte(byte(this.packetId >> 8))
	buffer2.WriteByte(byte(this.packetId & 0xFF))

	//Payload
	for i := 0; i < len(this.topics); i++ {
		topicLength := uint16(len(this.topics[i]))
		buffer2.WriteByte(byte(topicLength >> 8))
		buffer2.WriteByte(byte(topicLength & 0xFF))
		buffer2.WriteString(this.topics[i])
	}

	//2nd Pass

	//Fixed Header
	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buf2 := buffer2.Bytes()
	remainingLength := uint32(len(buf2))
	x, _ := this.EncodingRemainingLength(remainingLength)
	buffer.Write(x)

	//Viariable Header + Payload
	buffer.Write(buf2)

	return buffer.Bytes()
}

func (this *packet_unsubscribe) IParse(buffer []byte) error {
	var err error
	var bufferLength, remainingLength, consumedBytes, topicLength uint32

	bufferLength = uint32(len(buffer))

	if buffer == nil || bufferLength < 4+3 {
		return fmt.Errorf("Invalid %x Control Packet Size %x\n", this.packetType, bufferLength)
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
	if consumedBytes += 1; bufferLength < consumedBytes+remainingLength {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, remainingLength)
	}
	buffer = buffer[:consumedBytes+remainingLength]
	bufferLength = consumedBytes + remainingLength

	//Variable Header
	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	if consumedBytes += 2; bufferLength < consumedBytes+3 {
		return fmt.Errorf("Invalid %x Control Packet Must Have at least One Topic\n", this.packetType)
	}

	//Payload
	this.topics = nil
	for bufferLength > consumedBytes {
		if bufferLength < consumedBytes+3 {
			return fmt.Errorf("Invalid %x Control Packet Payload Length\n", this.packetType)
		}
		topicLength = ((uint32(buffer[consumedBytes])) << 8) | uint32(buffer[consumedBytes+1])
		if consumedBytes += 2; bufferLength < consumedBytes+topicLength || topicLength == 0 {
			return fmt.Errorf("Invalid %x Control Packet Topic Length\n", this.packetType)
		}

		this.topics = append(this.topics, string(buffer[consumedBytes:consumedBytes+topicLength]))
		consumedBytes += topicLength
	}

	return nil
}

//Variable Header
func (this *packet_unsubscribe) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_unsubscribe) SetPacketId(id uint16) {
	this.packetId = id
}

//Payload
func (this *packet_unsubscribe) GetUnsubscribeTopics() []string {
	return this.topics
}
func (this *packet_unsubscribe) SetUnsubscribeTopics(topics []string) {
	this.topics = topics
}
