package mqtt

import (
	"bytes"
	"errors"
	"fmt"
)

////////////////////Interface//////////////////////////////
type DisconnectPacket interface {
	Packet
}

////////////////////Implementation////////////////////////

type packet_disconnect struct {
	packet
}

func NewPacketDisconnect() *packet_disconnect {
	this := packet_disconnect{}
	this.IBytizer = &this
	this.IParser = &this
	return &this
}

func (this *packet_disconnect) IBytize() []byte {
	var buffer bytes.Buffer

	buffer.WriteByte((byte(this.packetType) << 4) | (this.packetFlag & 0x0F))
	buffer.WriteByte(0)

	return buffer.Bytes()
}

func (this *packet_disconnect) IParse(buffer []byte) error {
	if buffer == nil || len(buffer) != 2 {
		return errors.New("Invalid Control Packet Size")
	}

	if this.packetType = PacketType((buffer[0] >> 4) & 0x0F); this.packetType != PACKET_DISCONNECT {
		return fmt.Errorf("Invalid Control Packet Type %d\n", this.packetType)
	}
	if this.packetFlag = buffer[0] & 0x0F; this.packetFlag != 0 {
		return fmt.Errorf("Invalid Control Packet Flags %d\n", this.packetFlag)
	}
	if buffer[1] != 0 {
		return fmt.Errorf("Invalid Control Packet Remaining Length %d\n", buffer[1])
	}

	return nil
}
