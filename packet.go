package mqtt

type PacketType byte

const (
	PACKET_RESERVED_0 PacketType = iota
	PACKET_CONNECT
	PACKET_CONNACK
	PACKET_PUBLISH
	PACKET_PUBACK
	PACKET_PUBREC
	PACKET_PUBREL
	PACKET_PUBCOMP
	PACKET_SUBSCRIBE
	PACKET_SUBACK
	PACKET_UNSUBSCRIBE
	PACKET_PINGREQ
	PACKET_PINGRESP
	PACKET_DISCONNECT
	PACKET_RESERVED_15
)

type Packet interface {
	//Fixed Header
	GetType() PacketType
	SetType(pt PacketType)

	GetFlags() byte
	SetFlags(flags byte)
}

type ConnectPacket interface {
	Packet

	//Variable Header
	GetProtocalName() string
	SetProtocalName(n string)

	GetProtocalLevel() byte
	SetProtocalLevel(l byte)

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

type ConnackPacket interface {
	Packet

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() byte
	SetReturnCode(c byte)
}

type PublishPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetMessage() Message
	SetMessage(Message)
}

type PubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubrecPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubrelPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PubcompPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type SubscribePacket interface {
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

type SubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetReturnCode() []byte
	SetReturnCode([]byte)
}

type UnsubscribePacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetUnsubscribeTopics() []string
	SetUnsubscribeTopics([]string)
}

type UnsubackPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)
}

type PingreqPacket interface {
	Packet
}

type PingrespPacket interface {
	Packet
}

type DisconnectPacket interface {
	Packet
}
