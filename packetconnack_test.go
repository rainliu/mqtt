package mqtt

import (
	"testing"
)

func TestPacketConnack(t *testing.T) {
	input := []byte{0x20, 0x02, 0x01, 0xF0}

	pkt := NewPacketConnack()
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

	invalids := [][]byte{{0x20}, {0x20, 0x02}, {0x20, 0x02, 0x01}, {0x20, 0x02, 0x01, 0xF0, 0x2f},
		{0xF0, 0x02, 0x01, 0xF0},
		{0x22, 0x02, 0x01, 0xF0},
		{0x20, 0x00, 0x01, 0xF0},
		{0x20, 0x02, 0x21, 0xF0}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		}
	}
}
