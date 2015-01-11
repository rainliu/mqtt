package mqtt

import (
	"testing"
)

func TestPacketSubscribe(t *testing.T) {
	input := []byte{0x82, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F, 0x62, 0x01}

	pkt := NewPacketSubscribe()
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

	invalids := [][]byte{{0x82},
		{0x82, 0x08},
		{0x82, 0x08, 0x01},
		{0x82, 0x08, 0x01, 0xF0},
		{0x82, 0x08, 0x01, 0xF0, 0x00},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x03},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F, 0x62, 0x90},
		{0xF2, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F, 0x62, 0x01},
		{0x80, 0x08, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F, 0x62, 0x01},
		{0x82, 0x02, 0x01, 0xF0, 0x00, 0x03, 0x61, 0x2F, 0x62, 0x01},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x04, 0x61, 0x2F, 0x62, 0x01},
		{0x82, 0x08, 0x01, 0xF0, 0x00, 0x05, 0x61, 0x2F, 0x62, 0x01}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		} else {
			t.Logf("%v", invalids[i])
		}
	}
}
