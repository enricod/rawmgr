package common

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func data1() []byte {
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
	return data
}

func TestDecodeHuffTree(t *testing.T) {
	assert := assert.New(t)

	data := data1()

	v0, offset := ReadUint16Order(data, 0x9999, 0)
	assert.Equal(offset, int64(2), "offset must be 2")
	assert.Equal(uint16(0xffc4), v0, "marker")

	huffItems := GetHuffItems(data, 5)
	assert.Equal(16, len(huffItems), "nr elements ")

	for i := 0; i < 16; i++ {
		assert.Equal(i+1, huffItems[i].BitLength, "")
		assert.Equal(int(data[5+i]), huffItems[i].Count, "")
		assert.Equal(huffItems[i].Count, len(huffItems[i].Codes), "")
	}

	totValues := 0
	for i := 0; i < 16; i++ {
		totValues += huffItems[i].Count
	}

	assert.Equal(totValues, 16, "")
	assert.Equal(huffItems[1].Count, 1, "")
	assert.Equal(huffItems[1].Codes[0], uint8(0x00))
	assert.Equal(int(0x03), int(huffItems[2].Codes[2]), "3 bytes, the 3rd value is = 3")

	huffMappings := DecodeHuffTree(data)
	log.Printf("%v", huffMappings)
}
