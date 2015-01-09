package mqtt

////////////////////Interface//////////////////////////////

type PublishPacket interface {
	Packet

	//Variable Header
	GetPacketId() uint16
	SetPacketId(id uint16)

	//Payload
	GetMessage() Message
	SetMessage(Message)
}

////////////////////Implementation////////////////////////
