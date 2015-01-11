package mqtt

import (
	"bytes"
	"fmt"
)

////////////////////Interface//////////////////////////////
const (
	CONNECT_FLAG_RESERVED byte = 1 << iota
	CONNECT_FLAG_CLEAN_SESSION
	CONNECT_FLAG_WILL_FLAG
	CONNECT_FLAG_WILL_QOS
	CONNECT_FLAG_WILL_RETAIN
	CONNECT_FLAG_PASSWORD_FLAG
	CONNECT_FLAG_USERNAME_FLAG
)

type PacketConnect interface {
	Packet

	//Variable Header
	GetProtocolName() string
	SetProtocolName(n string)

	GetProtocolLevel() byte
	SetProtocolLevel(l byte)

	GetConnectFlags() byte
	SetConnectFlags(f byte)

	GetKeepAlive() uint16
	SetKeepAlive(t uint16)

	//Payload
	GetClientId() string
	SetClientId(s string)

	GetWillTopic() string
	SetWillTopic(s string)

	GetWillMessage() string
	SetWillMessage(s string)

	GetUserName() string
	SetUserName(s string)

	GetPassword() []byte
	SetPassword(s []byte)
}

////////////////////Implementation////////////////////////

type packet_connect struct {
	packet

	//Variable Header
	protocolName  string
	protocolLevel byte
	connectFlags  byte
	keepAlive     uint16

	//Payload
	clientId    string
	willTopic   string
	willMessage string
	userName    string
	password    []byte
}

func NewPacketConnect() *packet_connect {
	this := packet_connect{}
	this.IBytizer = &this
	this.IParser = &this
	return &this
}

func (this *packet_connect) IBytize() []byte {
	var buffer bytes.Buffer
	var buffer2 bytes.Buffer

	//1st Pass

	//Variable Header
	protocolLength := uint16(len(this.protocolName))
	buffer2.WriteByte(byte(protocolLength >> 8))
	buffer2.WriteByte(byte(protocolLength & 0xFF))
	buffer2.WriteString(this.protocolName)

	buffer2.WriteByte(this.protocolLevel)

	buffer2.WriteByte(this.connectFlags)

	buffer2.WriteByte(byte(this.keepAlive >> 8))
	buffer2.WriteByte(byte(this.keepAlive & 0xFF))

	//Payload
	clientId := this.EncodingUTF8(this.clientId)
	buffer2.Write(clientId)

	//Will Flag bit 2
	if (this.connectFlags & CONNECT_FLAG_WILL_FLAG) != 0 {
		buffer2.Write(this.EncodingUTF8(this.willTopic))
		buffer2.Write(this.EncodingUTF8(this.willMessage))
	}

	//UserName Flag bit 7
	if (this.connectFlags & CONNECT_FLAG_USERNAME_FLAG) != 0 {
		buffer2.Write(this.EncodingUTF8(this.userName))
	}

	//Password Flag bit 6
	if (this.connectFlags & CONNECT_FLAG_PASSWORD_FLAG) != 0 {
		buffer2.Write(this.EncodingBinary(this.password))
	}

	//2nd pass

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

func (this *packet_connect) IParse(buffer []byte) error {
	var err error
	var bufferLength, remainingLength, consumedBytes, utf8Bytes uint32

	bufferLength = uint32(len(buffer))

	if buffer == nil || bufferLength < 12 {
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
	protocolLength := ((uint32(buffer[consumedBytes])) << 8) | uint32(buffer[consumedBytes+1])
	consumedBytes += 2
	this.protocolName = string(buffer[consumedBytes : consumedBytes+protocolLength])
	consumedBytes += protocolLength

	this.protocolLevel = buffer[consumedBytes]
	consumedBytes += 1

	this.connectFlags = buffer[consumedBytes]
	consumedBytes += 1

	this.keepAlive = ((uint16(buffer[consumedBytes])) << 8) | uint16(buffer[consumedBytes+1])
	consumedBytes += 2

	//Payload
	if this.clientId, utf8Bytes, err = this.DecodingUTF8(buffer[consumedBytes:]); err != nil {
		return err
	}
	consumedBytes += utf8Bytes

	//Will Flag bit 2
	if (this.connectFlags & CONNECT_FLAG_WILL_FLAG) != 0 {
		if this.willTopic, utf8Bytes, err = this.DecodingUTF8(buffer[consumedBytes:]); err != nil {
			return err
		}
		consumedBytes += utf8Bytes

		if this.willMessage, utf8Bytes, err = this.DecodingUTF8(buffer[consumedBytes:]); err != nil {
			return err
		}
		consumedBytes += utf8Bytes
	}

	//UserName Flag bit 7
	if (this.connectFlags & CONNECT_FLAG_USERNAME_FLAG) != 0 {
		if this.userName, utf8Bytes, err = this.DecodingUTF8(buffer[consumedBytes:]); err != nil {
			return err
		}
		consumedBytes += utf8Bytes
	}

	//Password Flag bit 6
	if (this.connectFlags & CONNECT_FLAG_PASSWORD_FLAG) != 0 {
		if this.password, utf8Bytes, err = this.DecodingBinary(buffer[consumedBytes:]); err != nil {
			return err
		}
		consumedBytes += utf8Bytes
	}

	return nil
}

//Variable Header
func (this *packet_connect) GetProtocolName() string {
	return this.protocolName
}
func (this *packet_connect) SetProtocolName(n string) {
	this.protocolName = n
}

func (this *packet_connect) GetProtocolLevel() byte {
	return this.protocolLevel
}
func (this *packet_connect) SetProtocolLevel(l byte) {
	this.protocolLevel = l
}

func (this *packet_connect) GetConnectFlags() byte {
	return this.connectFlags
}
func (this *packet_connect) SetConnectFlags(f byte) {
	this.connectFlags = f
}

func (this *packet_connect) GetKeepAlive() uint16 {
	return this.keepAlive
}
func (this *packet_connect) SetKeepAlive(t uint16) {
	this.keepAlive = t
}

//Payload
func (this *packet_connect) GetClientId() string {
	return this.clientId
}
func (this *packet_connect) SetClientId(s string) {
	this.clientId = s
}

func (this *packet_connect) GetWillTopic() string {
	return this.willTopic
}
func (this *packet_connect) SetWillTopic(s string) {
	this.willTopic = s
}

func (this *packet_connect) GetWillMessage() string {
	return this.willMessage
}
func (this *packet_connect) SetWillMessage(s string) {
	this.willMessage = s
}

func (this *packet_connect) GetUserName() string {
	return this.userName
}
func (this *packet_connect) SetUserName(s string) {
	this.userName = s
}

func (this *packet_connect) GetPassword() []byte {
	return this.password
}
func (this *packet_connect) SetPassword(s []byte) {
	this.password = s
}
