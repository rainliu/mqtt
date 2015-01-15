package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////

type PacketAck interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PacketPuback interface {
	PacketAck
}

type PacketPubrec interface {
	PacketAck
}

type PacketPubrel interface {
	PacketAck
}

type PacketPubcomp interface {
	PacketAck
}

type PacketUnsuback interface {
	PacketAck
}

type PacketSuback interface {
	PacketAck

	//Payload
	GetReturnCodes() []byte
	SetReturnCodes([]byte)
}

type PacketConnack interface {
	PacketAck

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() CONNACK_RETURNCODE
	SetReturnCode(c CONNACK_RETURNCODE)
}

////////////////////Implementation////////////////////////

type packet_ack struct {
	packet

	packetId uint16
}

func NewPacketAcks(pt PacketType) *packet_ack {
	if !(pt == PACKET_PUBACK || pt == PACKET_PUBREC || pt == PACKET_PUBREL || pt == PACKET_PUBCOMP || pt == PACKET_UNSUBACK) {
		return nil
	}

	this := packet_ack{}

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

func (this *packet_ack) IBytize() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(2)
	buffer.WriteByte(byte(this.packetId >> 8))
	buffer.WriteByte(byte(this.packetId & 0xFF))

	return buffer.Bytes()
}

func (this *packet_ack) IParse(buffer []byte) error {
	if buffer == nil || len(buffer) != 4 {
		return fmt.Errorf("Invalid %s Control Packet Size %x\n", PACKET_TYPE_STRINGS[this.packetType], len(buffer))
	}

	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %s Control Packet Type %x\n", PACKET_TYPE_STRINGS[this.packetType], packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %s Control Packet Flags %x\n", PACKET_TYPE_STRINGS[this.packetType], packetFlag)
	}
	if buffer[1] != 2 {
		return fmt.Errorf("Invalid %s Control Packet Remaining Length %x\n", PACKET_TYPE_STRINGS[this.packetType], buffer[1])
	}

	if this.packetId = ((uint16(buffer[2])) << 8) | uint16(buffer[3]); this.packetId == 0 {
		return fmt.Errorf("Invalid %s Control Packet PacketId 0\n", PACKET_TYPE_STRINGS[this.packetType])
	}

	return nil
}

//Variable Header
func (this *packet_ack) GetPacketId() uint16 {
	return this.packetId
}
func (this *packet_ack) SetPacketId(id uint16) {
	this.packetId = id
}

/////////////////////
type packet_suback struct {
	packet_ack

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
		return fmt.Errorf("Invalid %x Control Packet Size %x\n", PACKET_TYPE_STRINGS[this.packetType], bufferLength)
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
	if consumedBytes += 2; bufferLength < consumedBytes+1 {
		return fmt.Errorf("Invalid %s Control Packet Must Have at least One Return Code\n", PACKET_TYPE_STRINGS[this.packetType])
	}

	//Payload
	this.returnCodes = make([]byte, remainingLength-2)
	copy(this.returnCodes, buffer[consumedBytes:consumedBytes+remainingLength-2])
	for i := 0; i < int(remainingLength-2); i++ {
		if !(this.returnCodes[i] <= 0x02 || this.returnCodes[i] == 0x80) {
			return fmt.Errorf("Invalid %s Control Packet Return Code %02x\n", PACKET_TYPE_STRINGS[this.packetType], this.returnCodes[i])
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
type CONNACK_RETURNCODE byte

const (
	CONNACK_RETURNCODE_ACCEPTED CONNACK_RETURNCODE = iota
	CONNACK_RETURNCODE_REFUSED_UNACCEPTABLE_PROTOCOL_VERSION
	CONNACK_RETURNCODE_REFUSED_IDENTIFIER_REJECTED
	CONNACK_RETURNCODE_REFUSED_SERVER_UNAVAILABLE
	CONNACK_RETURNCODE_REFUSED_BAD_USERNAME_OR_PASSWORD
	CONNACK_RETURNCODE_REFUSED_NOT_AUTHORIZED
)

type packet_connack struct {
	packet_ack

	spFlag     byte
	returnCode CONNACK_RETURNCODE
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
	buffer.WriteByte(byte(this.returnCode))

	return buffer.Bytes()
}

func (this *packet_connack) IParse(buffer []byte) error {
	if buffer == nil || len(buffer) != 4 {
		return fmt.Errorf("Invalid %s Control Packet Size %x\n", PACKET_TYPE_STRINGS[this.packetType], len(buffer))
	}

	//Fixed Header
	if packetType := PacketType((buffer[0] >> 4) & 0x0F); packetType != this.packetType {
		return fmt.Errorf("Invalid %s Control Packet Type %x\n", PACKET_TYPE_STRINGS[this.packetType], packetType)
	}
	if packetFlag := buffer[0] & 0x0F; packetFlag != this.packetFlag {
		return fmt.Errorf("Invalid %s Control Packet Flags %x\n", PACKET_TYPE_STRINGS[this.packetType], packetFlag)
	}
	if buffer[1] != 2 {
		return fmt.Errorf("Invalid %s Control Packet Remaining Length %x\n", PACKET_TYPE_STRINGS[this.packetType], buffer[1])
	}

	//Variable Header
	if buffer[2]&0xFE != 0 {
		return fmt.Errorf("Invalid %s Control Packet Acknowledge Flags %x\n", PACKET_TYPE_STRINGS[this.packetType], buffer[2])
	}
	this.spFlag = buffer[2] & 0x01

	if buffer[3] > 0x05 {
		return fmt.Errorf("Invalid %s Control Packet Return Code %x\n", PACKET_TYPE_STRINGS[this.packetType], buffer[3])
	}
	this.returnCode = CONNACK_RETURNCODE(buffer[3])

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

func (this *packet_connack) GetReturnCode() CONNACK_RETURNCODE {
	return this.returnCode
}
func (this *packet_connack) SetReturnCode(c CONNACK_RETURNCODE) {
	this.returnCode = c
}
