package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketPublish interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetMessage() Message
	SetMessage(Message)
}

////////////////////Implementation////////////////////////

type packet_publish struct {
	packet

	remainingLength uint32
	packetId        uint16
	message         Message
}

func NewPacketPublish() *packet_publish {
	this := packet_publish{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = PACKET_PUBLISH
	this.packetFlag = 0

	return &this
}

func (this *packet_publish) IBytize() []byte {
	var buffer bytes.Buffer
	var buffer2 bytes.Buffer

	//1st Pass

	//Variable Header
	topicLength := uint16(len(this.message.GetTopic()))
	buffer2.WriteByte(byte(topicLength >> 8))
	buffer2.WriteByte(byte(topicLength & 0xFF))
	buffer2.WriteString(this.message.GetTopic())
	buffer2.WriteByte(byte(this.packetId >> 8))
	buffer2.WriteByte(byte(this.packetId & 0xFF))

	//Payload
	buffer2.WriteString(this.message.GetContent())

	//2nd Pass

	//Fixed Header
	this.packetFlag = 0
	if this.message.GetDup() {
		this.packetFlag |= 0x08
	}
	this.packetFlag |= byte(this.message.GetQos()) << 1
	if this.message.GetRetain() {
		this.packetFlag |= 0x01
	}

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buf2 := buffer2.Bytes()
	this.remainingLength = uint32(len(buf2))
	x, _ := this.EncodingRemainingLength(this.remainingLength)
	buffer.Write(x)

	//Viariable Header + Payload
	buffer.Write(buf2)

	return buffer.Bytes()
}

func (this *packet_publish) IParse(buffer []byte) error {
	var err error
	var bufferLength uint32
	var consumedBytes uint32

	var dup bool
	var qos QOS
	var retain bool
	var topic string
	var content string

	bufferLength = uint32(len(buffer))
	if buffer == nil || bufferLength < 4 {
		return errors.New("Invalid Control Packet Size")
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid Control Packet Type %d\n", packetType)
	}
	this.packetFlag = buffer[0] & 0x0F
	if (buffer[0]>>1)&0x03 == 0x03 {
		return errors.New("Invalid Control Packet QoS level")
	} else {
		qos = QOS((buffer[0] >> 1) & 0x03)
	}
	if (buffer[0]>>3)&0x01 == 0x01 {
		dup = true
	} else {
		dup = false
	}
	if (buffer[0] & 0x01) == 0x01 {
		retain = true
	} else {
		retain = false
	}

	if this.remainingLength, consumedBytes, err = this.DecodingRemainingLength(buffer[1:]); err != nil {
		return err
	}
	if consumedBytes += 1; bufferLength < consumedBytes+this.remainingLength {
		return errors.New("Invalid Control Packet Size")
	}

	//Variable Header
	topicLength := ((uint32(buffer[consumedBytes])) << 8) | uint32(buffer[consumedBytes+1])
	if consumedBytes += 2; bufferLength < consumedBytes+topicLength {
		return errors.New("Invalid Control Packet Size")
	}

	topic = string(buffer[consumedBytes : consumedBytes+topicLength])
	if consumedBytes += topicLength; bufferLength < consumedBytes+2 {
		return errors.New("Invalid Control Packet Size")
	}

	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	if consumedBytes += 2; bufferLength < consumedBytes {
		return errors.New("Invalid Control Packet Size")
	}

	//Payload
	content = string(buffer[consumedBytes:])

	this.message = NewMessage(dup, qos, retain, topic, content)

	return nil
}

//Variable Header
func (this *packet_publish) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_publish) SetPacketId(id uint16) {
	this.packetId = id
}

//Payload
func (this *packet_publish) GetMessage() Message {
	return this.message
}
func (this *packet_publish) SetMessage(m Message) {
	this.message = m
}
