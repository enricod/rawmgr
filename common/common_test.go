package common

import (
	"encoding/binary"
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

		// data start #0
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

		// data start #1
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
		byte(0x00),

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

func TestShiftBytes(t *testing.T) {
	assert := assert.New(t)

	data := data1()
	huffMappings := DecodeHuffTree(data)

	huffMapping0 := huffMappings[0]

	assert.Equal(16, len(huffMapping0), "")

	mybytes := []byte{0xff, 0x00, 0xff, 0xfc, 0x00, 0xef, 0xf7, 0x00}
	num := binary.BigEndian.Uint64(mybytes)
	assert.Equal(uint64(0xff00fffc00eff700), num, "")

}

func TestHuffFindMapping(t *testing.T) {
	assert := assert.New(t)
	data := data1()

	huffMappings := DecodeHuffTree(data)
	huffMapping0 := huffMappings[0]

	m, err := HuffGetMapping(huffMapping0, 1022)
	assert.Nil(err)
	assert.Equal(0x0C, int(m.Value), "")

	m, err = HuffGetMapping(huffMapping0, 2)
	assert.Nil(err)
	assert.Equal(0x01, int(m.Value), "")

	m, err = HuffGetMapping(huffMapping0, 6)
	assert.Nil(err)
	assert.Equal(0x05, int(m.Value), "")

	m, err = HuffGetMapping(huffMapping0, 30)
	assert.Nil(err)
	assert.Equal(0x07, int(m.Value), "")

	m, err = HuffGetMapping(huffMapping0, 8190)
	assert.Nil(err)
	assert.Equal(0x0f, int(m.Value), "")

	// not found in mapping table
	m, err = HuffGetMapping(huffMapping0, 8062)
	assert.NotNil(err)

}
func TestGetHuffItems(t *testing.T) {
	assert := assert.New(t)

	data := data1()

	v0, offset := ReadUint16Order(data, 0x9999, 0)
	assert.Equal(offset, int64(2), "offset must be 2")
	assert.Equal(uint16(0xffc4), v0, "marker")

	huffItems, offset := GetHuffItems(data, 5)
	assert.Equal(16, len(huffItems), "nr elements ")

	huffItems1, _ := GetHuffItems(data, 5+33)
	assert.Equal(16, len(huffItems1), "nr elements ")

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

	// 13 bytes
	assert.Equal(1, int(huffItems[12].Count), "13 bytes, 1 code")
	assert.Equal(int(0x0f), int(huffItems[12].Codes[0]), "13 bytes, 1 code, value of the code is 15")

	assert.Equal(0, int(huffItems[13].Count), "14 bytes, no codes")
	assert.Equal(0, int(huffItems[15].Count), "16 bytes, no codes")

}

func TestDecodeHuffTree(t *testing.T) {
	assert := assert.New(t)
	data := data1()

	huffMappings := DecodeHuffTree(data)
	assert.NotEmpty(huffMappings[0])
	assert.NotEmpty(huffMappings[1])

	huffMapping0 := huffMappings[0]
	huffMapping1 := huffMappings[1]
	// 2 bits
	assert.Equal(huffMapping0[0].BitCount, 2, "")
	assert.Equal(huffMapping0[0].Code, uint64(0), "")
	assert.Equal(huffMapping0[0].Value, uint8(0x00), "")

	// 3 bits, first value
	assert.Equal(huffMapping0[1].BitCount, 3, "")
	assert.Equal(huffMapping0[1].Code, uint64(2), "")
	assert.Equal(huffMapping0[1].Value, uint8(1), "")

	assert.Equal(3, huffMapping0[2].BitCount, "")
	assert.Equal(huffMapping0[2].Code, uint64(3), "")
	assert.Equal(huffMapping0[2].Value, uint8(2), "")

	assert.Equal(6, huffMapping0[8].BitCount, "")
	assert.Equal(uint64(62), huffMapping0[8].Code, "")
	assert.Equal(uint8(8), huffMapping0[8].Value, "")

	assert.Equal(13, huffMapping0[15].BitCount, "")
	assert.Equal(uint64(8190), huffMapping0[15].Code, "")
	assert.Equal(uint8(15), huffMapping0[15].Value, "")

	// 2 bits
	assert.Equal(huffMapping1[0].BitCount, 2, "")
	assert.Equal(uint64(0), huffMapping1[0].Code, "")
	assert.Equal(uint8(0x00), huffMapping1[0].Value, "")
}

func TestHuffDifferences(t *testing.T) {
	assert := assert.New(t)

	diffes := HuffDifferences()
	assert.Equal(65535, len(diffes.Diffs), "")

	// 101110 = 46 base10
	d, _ := diffes.Find(14, uint16(46))
	assert.Equal(int32(-16337), d.Diff, "")

	d, _ = diffes.Find(13, uint16(1487))
	assert.Equal(int32(-6704), d.Diff, "")

	// 111 111 = 46 base10
	d, _ = diffes.Find(13, uint16(63))
	assert.Equal(int32(-8128), d.Diff, "")

	// 0000000001110
	d, _ = diffes.Find(13, uint16(14))
	assert.Equal(int32(-8177), d.Diff, "")

	// 15 = base10( 1111 )
	d, _ = diffes.Find(4, uint16(15))
	assert.Equal(int32(15), d.Diff, "")

	d, _ = diffes.Find(13, uint16(7168))
	assert.Equal(int32(7168), d.Diff, "")

	d, _ = diffes.Find(11, uint16(792))
	assert.Equal(int32(-1255), d.Diff, "")

	d, _ = diffes.Find(13, uint16(6767))
	assert.Equal(int32(6767), d.Diff, "")

	// 32767 = base10( 111111111111111 )
	d, _ = diffes.Find(15, uint16(32767))
	assert.Equal(int32(32767), d.Diff, "")

	d, _ = diffes.Find(1, uint16(1))
	assert.Equal(int32(1), d.Diff, "")

	// 32767 = base10( 111111111111111 )
	d, _ = diffes.Find(15, uint16(0))
	assert.Equal(int32(-32767), d.Diff, "")

	d, _ = diffes.Find(14, uint16(8440))
	assert.Equal(int32(8440), d.Diff, "")

	d, _ = diffes.Find(14, uint16(2047))
	assert.Equal(int32(-14336), d.Diff, "")

	d, _ = diffes.Find(13, uint16(1510))
	assert.Equal(int32(-6681), d.Diff, "")
}
