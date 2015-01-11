package mqtt

import (
	"testing"
)

func TestPacketSuback(t *testing.T) {
	input := []byte{0x90, 0x03, 0x01, 0xF0, 0x01}

	pkt := NewPacketSuback()
	if err := pkt.Parse(input); err != nil {
		t.Errorf(err.Error())
		return
	}

	output := pkt.Bytes()
	if len(input) == len(output) {
		for i := 0; i < len(input); i++ {
			if input[i] != output[i] {
				t.Errorf("Mismatch %02x vs %02x\n", input[i], output[i])
			}
		}
	} else {
		t.Errorf("Mismatch length %x vs %x\n", len(input), len(output))
	}

	invalids := [][]byte{{0x90},
		{0x90, 0x03},
		{0x90, 0x03, 0x01},
		{0x90, 0x03, 0x01, 0xF0},
		{0x90, 0x03, 0x01, 0xF0, 0x01, 0x03},
		{0xF0, 0x03, 0x01, 0xF0, 0x01},
		{0x92, 0x03, 0x01, 0xF0, 0x80},
		{0x90, 0x02, 0x01, 0xF0, 0x00},
		{0x90, 0x03, 0x01, 0xF0, 0x81},
		{0x90, 0x03, 0x01, 0xF0, 0x03}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		} else {
			t.Logf("%v", invalids[i])
		}
	}
}
