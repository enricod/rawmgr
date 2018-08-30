package common

import (
	"testing"
)

func TestDecodeHuffTree(t *testing.T) {

	data := []byte{
		// 0xffc4
		byte(0xff),
		byte(0xc4),

		// length
		byte(0x00),
		byte(0x44),

		// table class / table index
		byte(0x00),

		// data start
		byte(0x00),
		byte(0x01),
		byte(0x05),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x00),
		byte(0x00),
		byte(0x00),

		// ---
		byte(0x00),
		byte(0x01),
		byte(0x02),
		byte(0x03),
		byte(0x04),
		byte(0x05),
		byte(0x06),
		byte(0x07),
		byte(0x08),
		byte(0x09),
		byte(0x0a),
		byte(0x0b),
		byte(0x0c),
		byte(0x0d),
		byte(0x0e),
		byte(0x0f),

		// ---- Table class / huffmant table index
		byte(0x01),

		// ---
		byte(0x00),
		byte(0x03),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),
		byte(0x01),

		// -- valori
		byte(0x00),
		byte(0x01),
		byte(0x02),
		byte(0x03),
		byte(0x04),
		byte(0x05),
		byte(0x06),
		byte(0x07),
		byte(0x08),
		byte(0x09),
		byte(0x0a),
		byte(0x0b),
		byte(0x0c),
		byte(0x0d),
		byte(0x0e),
		byte(0x0f),

		// -- start SOF3 Header
		byte(0xff),
		byte(0xc3),
	}

	v0, offset := ReadUint16Order(data, 0x9999, 0)
	if offset != 2 {
		t.Errorf("offset deve essere 2")
	}
	if v0 != 0xffc4 {
		t.Errorf("first value must be 0xffc4, found=%d", v0)
	}

	huffItems := GetHuffItems(data, 5)
	if 16 != len(huffItems) {
		t.Errorf("devono essere 16 elementi, found=%d", len(huffItems))
	}

	for i := 0; i < 16; i++ {
		if huffItems[i].BitLength != i+1 {
			t.Errorf("BitLength deve essere %d, found=%d", int(data[5+i]), huffItems[i].BitLength)
		}
		if huffItems[i].Count != int(data[5+i]) {
			t.Errorf("Count deve essere %d, found=%d", int(data[5+i]), huffItems[i].Count)
		}
		if huffItems[i].Count != len(huffItems[i].Codes) {
			t.Errorf("Items count, expected %d, got %d", huffItems[i].Count, len(huffItems[i].Codes))
		}
	}

	totValues := 0
	for i := 0; i < 16; i++ {
		totValues += huffItems[i].Count
	}

	if totValues != 16 {
		t.Errorf("totValues deve essere 16, found=%d", totValues)
	}
}
