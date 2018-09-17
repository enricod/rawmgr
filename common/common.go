package common

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
)

// Verbose true if you want more output
var Verbose *bool

// LittleEndian value
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

// ReadUint32 reads 4 bytes, coverts to uint32
func ReadUint32(data []byte, order uint16, offset int64) (uint32, int64) {
	var b1 = data[offset : offset+4]
	return binary.BigEndian.Uint32(b1), offset + 4
}

// GetUint16 a partire da offset legge un int uint16 e torna la nuova posizione
func GetUint16(f *os.File, offset int64) (uint16, int64) {
	value := binary.BigEndian.Uint16(readFromFileBytes(f, offset, 2))
	return value, offset + 2
}

// GetUint32 reads 4 bytes, coverts to uint32
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

// Get1Byte reads from file 1 byte and returns an uint16
func Get1Byte(f *os.File, offset int64) (uint16, int64) {
	mybyte := readFromFileBytes(f, offset, 1)[0]
	return uint16(mybyte), offset + 1
}

// GetInt reads from the file 2 bytes ... incomplete and undocumented, to be deleted
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

// HuffItem item in huffman tree
type HuffItem struct {
	BitLength int
	Count     int
	Codes     []byte
}

// GetHuffItems from the file extracts the data for the Huffman tree
// the first 16 bytes give the number of bytes to associate to the item, which are taken from an array starting offset+16
//
//               offset
//  |----|----|----|
//   ffc4 0042 0000
//
//  |----|----|----|----|----|----|----|----
//   0104 0203 0101 0101 0100 0000 0000 00
//
//
// returns something like
// [
//	 {1 bit, []}
//   {2 bit, [ 0 ] }
//   {3 bit, [1,2,3,4,5] }
//      ...
// ]
func GetHuffItems(data []byte, offset int64) ([]HuffItem, int64) {
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
		someCodes := vals[0:item.Count]
		vals = vals[item.Count:]
		huffItems[i] = HuffItem{Count: item.Count, BitLength: item.BitLength, Codes: someCodes}
	}
	return huffItems, offset + int64(16+totValues)
}

// NSpaces builds a string of 'spaces' spaces
func NSpaces(spaces int) string {
	var buffer bytes.Buffer
	for i := 0; i < spaces; i++ {
		buffer.WriteString("    ")
	}
	return buffer.String()
}

type HuffMappingKey struct {
	BitCount int
	Code     uint64
}

// HuffMapping mapping between a bitq sequence and a byte, used in Huffman decoding
type HuffMapping struct {
	BitCount int
	Value    byte
	Code     uint64
}

func removeFromNextLine(lines [][]uint64, row int, howmany int) {
	if row == len(lines)-1 {
		return
	}
	lines[row+1] = lines[row+1][howmany:len(lines[row+1])]
	removeFromNextLine(lines, row+1, howmany*2)
}

func decodeHuff(huffItems []HuffItem) []HuffMapping {
	valuesPerBitsNum := make([][]uint64, 16)
	for i := range valuesPerBitsNum {
		valuesPerBitsNum[i] = make([]uint64, uint32(math.Pow(2, float64(i+1))))
		for j := range valuesPerBitsNum[i] {
			valuesPerBitsNum[i][j] = uint64(j)
		}
	}
	codesMapping := []HuffMapping{}

	for i := range huffItems {
		for j := range huffItems[i].Codes {
			//log.Printf("%d, %d, %v", i, j, huffIems0[i])
			codesMapping = append(codesMapping, HuffMapping{BitCount: huffItems[i].BitLength, Value: huffItems[i].Codes[j], Code: valuesPerBitsNum[i][j]})
			//log.Printf("%d, %d, %v => %v", i, j, huffIems0[i], codesMapping)
			removeFromNextLine(valuesPerBitsNum, i, 2)
		}
	}
	return codesMapping
}

// DecodeHuffTree builds Huffman table (starts at the 5th byte in header)
func DecodeHuffTree(data []byte) [][]HuffMapping {

	result := [][]HuffMapping{}
	huffIems0, offset := GetHuffItems(data, 5)
	result = append(result, decodeHuff(huffIems0))

	// fixme
	huffItems1, _ := GetHuffItems(data, offset+1)
	result = append(result, decodeHuff(huffItems1))

	log.Printf("%v", huffIems0)
	return result
}

// PopFirst extracts first byte from array
func PopFirst(s []byte) (byte, []byte) {
	first := s[0]
	copy(s, s[1:])
	s2 := s[:len(s)-1]
	return first, s2
}

// Pow2 return 2^exp as uint64
func Pow2(exp int) uint64 {
	return uint64(math.Pow(2, float64(exp)))
}

// HuffGetMapping given the mappings and a values, return the mapping matching, error if not found
func HuffGetMapping(huffMappings []HuffMapping, code uint64, bitsLenght int) (HuffMapping, error) {

	for i := len(huffMappings) - 1; i >= 0; i-- {
		h := huffMappings[i]
		if h.BitCount == bitsLenght && h.Code == code {
			return h, nil
		}
	}
	return HuffMapping{}, fmt.Errorf("code not found %d", code)
}

type HuffDiff struct {
	BitCount uint8
	Key      uint16
	Diff     int32
}

type HuffDiffs struct {
	Diffs []HuffDiff
}

func (r *HuffDiffs) Find(bitCount uint8, key uint16) (HuffDiff, error) {
	for _, d := range r.Diffs {
		if d.BitCount == bitCount && d.Key == key {
			return d, nil
		}
	}
	return HuffDiff{}, fmt.Errorf("not found")
}

func HuffDifferences() HuffDiffs {
	result := []HuffDiff{}
	result = append(result, HuffDiff{0, 0, 0})
	for nbits := 1; nbits < 16; nbits++ {
		tot := Pow2(nbits)
		diff := -1 * int(tot-1)
		key := 0
		for v := 0; v < int(tot)/2; v++ {
			result = append(result, HuffDiff{BitCount: uint8(nbits), Key: uint16(key), Diff: int32(diff + v)})
			key++
		}
		diff = int(tot - tot/2)
		for v := 0; v < int(tot)/2; v++ {
			result = append(result, HuffDiff{BitCount: uint8(nbits), Key: uint16(key), Diff: int32(diff + v)})
			key++
		}

	}
	return HuffDiffs{result}
}
