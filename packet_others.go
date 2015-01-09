package mqtt

////////////////////Interface//////////////////////////////
type PingreqPacket interface {
	Packet
}

type PingrespPacket interface {
	Packet
}

type DisconnectPacket interface {
	Packet
}

////////////////////Implementation////////////////////////

type packet_pingreq packet

type packet_pingresp packet

type packet_disconnect packet
