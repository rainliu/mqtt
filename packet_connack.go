package mqtt

////////////////////Interface//////////////////////////////

type ConnackPacket interface {
	Packet

	//Variable Header
	GetSPFlag() bool
	SetSPFlag(b bool)

	GetReturnCode() byte
	SetReturnCode(c byte)
}

////////////////////Implementation////////////////////////
