package common

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

// Verbose true if you want more output
var Verbose *bool

const LittleEndian = 0x4949

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// ReadUint16 reads 2 byte in offset index and converts in uint16 (BigEndian), returning also new array index
func ReadUint16(data []byte, offset int64) (uint16, int64) {
	var b1 = data[offset : offset+2]
	return binary.BigEndian.Uint16(b1), offset + 2
}

// ReadUint16Order reads 2 byte in offset index and converts in uint16 (byte order spoecified), returning also new array index
func ReadUint16Order(data []byte, order uint16, offset int64) (uint16, int64) {
	var b1 = data[offset : offset+2]
	if order == LittleEndian {
		// little endian
		return binary.LittleEndian.Uint16(b1), offset + 2
	}
	return binary.BigEndian.Uint16(b1), offset + 2
}

func ReadUint32Order(data []byte, order uint16, offset int64) (uint32, int64) {
	var b1 = data[offset : offset+4]
	if order == LittleEndian {
		return binary.LittleEndian.Uint32(b1), offset + 4
	}
	return binary.BigEndian.Uint32(b1), offset + 4
}

// ReadUint8 reads 1 byte, coverts to uint8
func ReadUint8(data []byte, offset int64) (uint8, int64) {
	return uint8(data[offset]), offset + 1
}

func ReadUint32(data []byte, order uint16, offset int64) (uint32, int64) {
	var b1 = data[offset : offset+4]
	return binary.BigEndian.Uint32(b1), offset + 4
}

// GetUint16 a partire da offset legge un int uint16 e torna la nuova posizione
func GetUint16(f *os.File, offset int64) (uint16, int64) {
	value := binary.BigEndian.Uint16(readFromFileBytes(f, offset, 2))
	return value, offset + 2
}

func GetUint32(f *os.File, offset int64) (uint32, int64) {
	value := binary.BigEndian.Uint32(readFromFileBytes(f, offset, 4))
	return value, offset + 4
}

func GetUint16WithOrder(f *os.File, order uint16, offset int64) (uint16, int64) {
	var value uint16
	if order == 0x4949 {
		// little endian
		value = binary.LittleEndian.Uint16(readFromFileBytes(f, offset, 2))
	} else {
		value = binary.BigEndian.Uint16(readFromFileBytes(f, offset, 2))
	}
	return value, offset + 2
}

func GetUint32WithOrder(f *os.File, order uint16, offset int64) (uint32, int64) {
	var value uint32
	if order == 0x4949 {
		// little endian
		value = binary.LittleEndian.Uint32(readFromFileBytes(f, offset, 4))
	} else {
		value = binary.BigEndian.Uint32(readFromFileBytes(f, offset, 4))
	}
	return value, offset + 4
}

func Get1Byte(f *os.File, offset int64) (uint16, int64) {
	mybyte := readFromFileBytes(f, offset, 1)[0]
	return uint16(mybyte), offset + 1
}

// GetInt
func GetInt(f *os.File, order uint16, offset int64, typ uint16) (int64, int64) {
	if typ == 3 {
		v, start := GetUint16WithOrder(f, order, offset)
		return int64(v), start
	}
	v, start := GetUint16WithOrder(f, order, offset)
	return int64(v), start
}

func readFromFileBytes(f *os.File, start int64, howmany int64) []byte {
	_, err := f.Seek(start, 0)
	check(err)
	retBytes := make([]byte, howmany)
	f.Read(retBytes)
	check(err)
	//log.Printf("lettura di %d bytes @ %d: %s\n", n2, o2, string(retBytes))
	return retBytes
}

type HuffCode struct {
	Value    byte
	BytesStr string
}
type HuffItem struct {
	BitLength int
	Count     int
	Codes     []HuffCode
}
type HuffTree struct {
	Left  *HuffItem
	Item  *HuffItem
	Right *HuffItem
}

func GetHuffItems(data []byte, offset int64) []HuffItem {
	//length, offset := ReadUint16Order(data, LittleEndian, 2)
	nrCodesOfLength := data[offset : offset+16]
	var huffItems []HuffItem
	totValues := 0
	for i := 0; i < 16; i++ {
		var item = HuffItem{BitLength: i + 1, Count: int(nrCodesOfLength[i])}
		huffItems = append(huffItems, item)
		totValues += item.Count
	}

	var vals []byte
	var item HuffItem
	vals = data[offset+16 : offset+16+int64(totValues)]
	for i := 0; i < 16; i++ {
		item = huffItems[i]
		codes := make([]HuffCode, 0)

		for j := 0; j < item.Count; j++ {
			var first byte
			first, vals = PopFirst(vals)
			var huffCode = HuffCode{Value: first}
			codes = append(codes, huffCode)
		}

		huffItems[i] = HuffItem{Count: item.Count, BitLength: item.BitLength, Codes: codes}
		fmt.Printf("numero di codes %d vs %d\n", len(item.Codes), len(codes))
	}
	return huffItems
}

// DecodeHuffTree build Huffman table
func DecodeHuffTree(data []byte) *HuffTree {
	log.Printf("huff data %v", data)

	var result HuffTree

	return &result
}

func PopFirst(s []byte) (byte, []byte) {
	first := s[0]
	copy(s, s[1:])
	s2 := s[:len(s)-1]
	return first, s2
}
