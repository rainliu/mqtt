package mqtt

import (
	"testing"
)

func TestPacketPingreq(t *testing.T) {
	input := []byte{0xC0, 0x00}

	pkt := NewPacket(PACKET_PINGREQ)
	if err := pkt.Parse(input); err != nil {
		t.Errorf(err.Error())
	}

	output := pkt.Bytes()
	for i := 0; i < len(input); i++ {
		if input[i] != output[i] {
			t.Errorf("Mismatch %02x vs %02x\n", input[i], output[i])
		}
	}

	invalids := [][]byte{{0xC0}, {0xC0, 0x00, 0x00}, {0xF0, 0x00}, {0xC2, 0x00}, {0xC0, 0x02}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		}
	}
}

func TestPacketPingresp(t *testing.T) {
	input := []byte{0xD0, 0x00}

	pkt := NewPacket(PACKET_PINGRESP)
	if err := pkt.Parse(input); err != nil {
		t.Errorf(err.Error())
	}

	output := pkt.Bytes()
	for i := 0; i < len(input); i++ {
		if input[i] != output[i] {
			t.Errorf("Mismatch %02x vs %02x\n", input[i], output[i])
		}
	}

	invalids := [][]byte{{0xD0}, {0xD0, 0x00, 0x00}, {0xF0, 0x00}, {0xD2, 0x00}, {0xD0, 0x02}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		}
	}
}

func TestPacketDisconnect(t *testing.T) {
	input := []byte{0xE0, 0x00}

	pkt := NewPacket(PACKET_DISCONNECT)
	if err := pkt.Parse(input); err != nil {
		t.Errorf(err.Error())
	}

	output := pkt.Bytes()
	for i := 0; i < len(input); i++ {
		if input[i] != output[i] {
			t.Errorf("Mismatch %02x vs %02x\n", input[i], output[i])
		}
	}

	invalids := [][]byte{{0xE0}, {0xE0, 0x00, 0x00}, {0xF0, 0x00}, {0xE2, 0x00}, {0xE0, 0x02}}
	for i := 0; i < len(invalids); i++ {
		if err := pkt.Parse(invalids[i]); err != nil {
			t.Logf(err.Error())
		}
	}
}
