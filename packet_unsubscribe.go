package mqtt

////////////////////Interface//////////////////////////////

type UnsubscribePacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetUnsubscribeTopics() []string
	SetUnsubscribeTopics([]string)
}

////////////////////Implementation////////////////////////
