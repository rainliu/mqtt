package mqtt

////////////////////Interface//////////////////////////////

type PacketPublish interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetMessage() Message
	SetMessage(Message)
}

////////////////////Implementation////////////////////////
