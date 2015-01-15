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

	GetQoSs() []QOS
	SetQoSs([]QOS)
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
		return fmt.Errorf("Invalid %s Control Packet Size %x\n", PACKET_TYPE_STRINGS[this.packetType], bufferLength)
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %s Control Packet Type %x\n", PACKET_TYPE_STRINGS[this.packetType], packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %s Control Packet Flags %x\n", PACKET_TYPE_STRINGS[this.packetType], packetFlag)
	}
	if remainingLength, consumedBytes, err = this.DecodingRemainingLength(buffer[1:]); err != nil {
		return fmt.Errorf("Invalid %s Control Packet DecodingRemainingLength %s\n", PACKET_TYPE_STRINGS[this.packetType], err.Error())
	}
	if consumedBytes += 1; bufferLength < consumedBytes+remainingLength {
		return fmt.Errorf("Invalid %s Control Packet Remaining Length %x\n", PACKET_TYPE_STRINGS[this.packetType], remainingLength)
	}
	buffer = buffer[:consumedBytes+remainingLength]
	bufferLength = consumedBytes + remainingLength

	//Variable Header
	if this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1]); this.packetId == 0 {
		return fmt.Errorf("Invalid %s Control Packet PacketId 0\n", PACKET_TYPE_STRINGS[this.packetType])
	}
	if consumedBytes += 2; bufferLength < consumedBytes+4 {
		return fmt.Errorf("Invalid %s Control Packet Payload Length\n", PACKET_TYPE_STRINGS[this.packetType])
	}

	//Payload
	this.topics = nil
	this.qos = nil
	for bufferLength > consumedBytes {
		topicLength = ((uint32(buffer[consumedBytes])) << 8) | uint32(buffer[consumedBytes+1])
		if consumedBytes += 2; bufferLength < consumedBytes+topicLength || topicLength == 0 {
			return fmt.Errorf("Invalid %s Control Packet Topic Length %x\n", PACKET_TYPE_STRINGS[this.packetType], topicLength)
		}

		this.topics = append(this.topics, string(buffer[consumedBytes:consumedBytes+topicLength]))
		if consumedBytes += topicLength; bufferLength < consumedBytes+1 {
			return fmt.Errorf("Invalid %s Control Packet QoS Length\n", PACKET_TYPE_STRINGS[this.packetType])
		}

		if buffer[consumedBytes] > 2 {
			return fmt.Errorf("Invalid %s Control Packet QoS Level\n", PACKET_TYPE_STRINGS[this.packetType])
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

func (this *packet_subscribe) GetQoSs() []QOS {
	return this.qos
}
func (this *packet_subscribe) SetQoSs(qos []QOS) {
	this.qos = qos
}
