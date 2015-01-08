package mqtt

import "bytes"

////////////////////Interface//////////////////////////////

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

	GetPassword() string
	SetPassword(s string)
}

////////////////////Implementation////////////////////////

type packet_connect struct {
	packet

	protocolName  string
	protocolLevel byte
	connectFlgs   byte
	keepAlive     uint16
	clientId      string
	willTopic     string
	willMessage   string
	userName      string
	password      string
}

func NewPacketConnect() *packet_connect {
	this := packet_connect{}
	this.IBytizer = &this
	this.IParser = &this
	return &this
}

func (this *packet_connect) IBytize() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(0)

	return buffer.Bytes()
}

func (this *packet_connect) IParse(buffer []byte) error {
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
	return this.connectFlgs
}
func (this *packet_connect) SetConnectFlags(f byte) {
	this.connectFlgs = f
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

func (this *packet_connect) GetPassword() string {
	return this.password
}
func (this *packet_connect) SetPassword(s string) {
	this.password = s
}
