package common

import (
	"encoding/binary"
	"os"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// getUint16 a partire da offset legge un int uint16 e torna la nuova posizione
func GetUint16(f *os.File, offset int64) (uint16, int64) {
	value := binary.BigEndian.Uint16(readFromFileBytes(f, offset, 2))
	return value, offset + 2
}

func GetUint32(f *os.File, offset int64) (uint32, int64) {
	value := binary.BigEndian.Uint32(readFromFileBytes(f, offset, 4))
	return value, offset + 4
}

func Get1Byte(f *os.File, offset int64) (uint16, int64) {
	mybyte := readFromFileBytes(f, offset, 1)[0]
	return uint16(mybyte), offset + 1
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