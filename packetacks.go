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

type PacketSuback interface {
	PacketAcks

	//Payload
	GetReturnCodes() []byte
	SetReturnCodes([]byte)
}

type PacketConnack interface {
	Packet

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() byte
	SetReturnCode(c byte)
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

/////////////////////
type packet_suback struct {
	packet_acks

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
	var bufferLength, remainingLength, consumedBytes uint32

	bufferLength = uint32(len(buffer))

	if buffer == nil || bufferLength < 5 {
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
		return fmt.Errorf("Invalid %x Control Packet DecodingRemainingLength %s\n", this.packetType, err.Error())
	}
	if consumedBytes += 1; bufferLength < consumedBytes+remainingLength {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, remainingLength)
	}
	buffer = buffer[:consumedBytes+remainingLength]
	bufferLength = consumedBytes + remainingLength

	//Variable Header
	this.packetId = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	if consumedBytes += 2; bufferLength < consumedBytes+1 {
		return fmt.Errorf("Invalid %x Control Packet Must Have at least One Return Code\n", this.packetType)
	}

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

//Payload
func (this *packet_suback) GetReturnCodes() []byte {
	return this.returnCodes
}
func (this *packet_suback) SetReturnCodes(returnCodes []byte) {
	this.returnCodes = make([]byte, len(returnCodes))
	copy(this.returnCodes, returnCodes)
}

////////////////////Implementation////////////////////////
const (
	CONNACK_RETURNCODE_ACCEPTED byte = iota
	CONNACK_RETURNCODE_REFUSED_UNACCEPTABLE_PROTOCOL_VERSION
	CONNACK_RETURNCODE_REFUSED_IDENTIFIER_REJECTED
	CONNACK_RETURNCODE_REFUSED_SERVER_UNAVAILABLE
	CONNACK_RETURNCODE_REFUSED_BAD_USERNAME_OR_PASSWORD
	CONNACK_RETURNCODE_REFUSED_NOT_AUTHORIZED
)

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
		return fmt.Errorf("Invalid %x Control Packet Size %x\n", this.packetType, len(buffer))
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %x Control Packet Type %x\n", this.packetType, packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %x Control Packet Flags %x\n", this.packetType, packetFlag)
	}
	if buffer[1] != 2 {
		return fmt.Errorf("Invalid %x Control Packet Remaining Length %x\n", this.packetType, buffer[1])
	}

	//Variable Header
	if buffer[2]&0xFE != 0 {
		return fmt.Errorf("Invalid %x Control Packet Acknowledge Flags %x\n", this.packetType, buffer[2])
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
