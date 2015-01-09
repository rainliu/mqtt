package mqtt

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
