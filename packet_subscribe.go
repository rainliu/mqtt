package mqtt

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
