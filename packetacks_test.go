package mqtt

import (
	"testing"
)

func TestPacketAcks(t *testing.T) {
	input := []byte{0xB0, 0x02, 0x0F, 0xF0}

	pkt := NewPacketAcks(PACKET_UNSUBACK)
	if err := pkt.Parse(input); err != nil {
		t.Errorf(err.Error())
		return
	}

	output := pkt.Bytes()
	for i := 0; i < len(input); i++ {
		if input[i] != output[i] {
			t.Errorf("Mismatch %02x vs %02x\n", input[i], output[i])
		}
	}

	invalids := [][]byte{{0xB0}, {0xB0, 0x02}, {0xB0, 0x02, 0x0F}, {0xB0, 0x02, 0x0F, 0xF0, 0x2f},
		{0xF0, 0x02, 0x0F, 0xF0},
		{0xB2, 0x02, 0x0F, 0xF0},
		{0xB0, 0x00, 0x0F, 0xF0}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		}
	}
}
