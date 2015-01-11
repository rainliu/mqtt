package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketSubscribe interface {
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

////////////////////Implementation////////////////////////

type packet_subscribe struct {
	packet

	packetId uint16
	topics   []string
	qos      []QOS
}

func NewPacketSubscribe() *packet_subscribe {
	this := packet_subscribe{}

	this.IBytizer = &this
	this.IParser = &this

	this.packetType = PACKET_SUBSCRIBE
	this.packetFlag = 2

	return &this
}

func (this *packet_subscribe) IBytize() []byte {
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
		buffer2.WriteByte(byte(this.qos[i]))
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

func (this *packet_subscribe) IParse(buffer []byte) error {
	var err error
	var bufferLength, remainingLength, consumedBytes, topicLength uint32

	bufferLength = uint32(len(buffer))

	if buffer == nil || bufferLength < 4+4 {
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
	consumedBytes += 1
	if bufferLength < consumedBytes+remainingLength {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, remainingLength)
	}

	//Variable Header
	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	consumedBytes += 2
	if bufferLength < consumedBytes+4 {
		return fmt.Errorf("Invalid %x Control Packet Payload Length\n", this.packetType)
	}

	//Payload
	this.topics = nil
	this.qos = nil
	for bufferLength > consumedBytes {
		topicLength = ((uint32(buffer[consumedBytes])) << 8) | uint32(buffer[consumedBytes+1])
		consumedBytes += 2
		if bufferLength < consumedBytes+topicLength {
			return fmt.Errorf("Invalid %x Control Packet Topic Length %x\n", this.packetType, topicLength)
		}

		this.topics = append(this.topics, string(buffer[consumedBytes:consumedBytes+topicLength]))
		consumedBytes += topicLength
		if bufferLength < consumedBytes+1 {
			return fmt.Errorf("Invalid %x Control Packet QoS Length\n", this.packetType)
		}

		if buffer[consumedBytes] > 2 {
			return fmt.Errorf("Invalid %x Control Packet QoS Level\n", this.packetType)
		}
		this.qos = append(this.qos, QOS(buffer[consumedBytes]))
		consumedBytes += 1
	}

	return nil
}

//Variable Header
func (this *packet_subscribe) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_subscribe) SetPacketId(id uint16) {
	this.packetId = id
}

//Payload
func (this *packet_subscribe) GetSubscribeTopics() []string {
	return this.topics
}
func (this *packet_subscribe) SetSubscribeTopics(topics []string) {
	this.topics = topics
}

func (this *packet_subscribe) GetQos() []QOS {
	return this.qos
}
func (this *packet_subscribe) SetQos(qos []QOS) {
	this.qos = qos
}
