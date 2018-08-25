package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/enricod/rawmgr/common"
)

var Tags = []map[uint16]string{{
	0x0100: "width",
	0x0101: "height",
	272:    "model",
	0x0111: "stripOffset",
	0x0112: "orientation",
	0x011a: "xResolution",
	0x8729: "exif",
},
	{
		0x829a: "exposureTime",
		0x829d: "fNumber",
	},
	{
		0x0001: "canonCameraSettings",
		0x0002: "canonFocalLength",
	},
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Header struct {
	ByteOrder       uint16
	TiffMagicWord   uint16
	IfdOffset       int64
	CR2MagicWord    string
	CR2MajorVersion uint8
	CR2MinorVersion uint8
	RawIfdOffset    int64
}

type IFD struct {
	Tag     uint16
	Typ     uint16
	Count   uint32
	Value   uint32
	Level   int
	SubIFDs IFDs
}

type IFDs struct {
	EntriesNr     uint16
	Offset        int64
	Ifds          []IFD
	NextIfdOffset int64
}

func readHeader(data []byte) (Header, error) {

	var start int64
	var result = Header{}

	// 18761 "II" or 0x4949 means Intel byte order (little endian)
	// "MM" or 0x4d4d means Motorola byte order (big endian)
	result.ByteOrder, start = common.ReadUint16(data, start)
	result.TiffMagicWord = uint16(data[start])
	if result.TiffMagicWord != 0x002A {
		return result, fmt.Errorf("TiffMagicWord not valid %d", result.TiffMagicWord)
	}

	var ifdOffset, rawIfdOffset uint16
	ifdOffset, start = common.ReadUint16Order(data, result.ByteOrder, 4)

	result.IfdOffset = int64(ifdOffset)
	result.CR2MagicWord = string(data[8:10])
	result.CR2MajorVersion, start = common.ReadUint8(data, 10)
	result.CR2MinorVersion, start = common.ReadUint8(data, start)
	rawIfdOffset, start = common.ReadUint16Order(data, result.ByteOrder, start)
	result.RawIfdOffset = int64(rawIfdOffset)

	return result, nil

}

// data is 12 bytes long
func readIfd(data []byte) IFD {
	var result = IFD{}

	result.Tag = binary.LittleEndian.Uint16(data[0:2])
	result.Typ = binary.LittleEndian.Uint16(data[2:4])
	result.Count = binary.LittleEndian.Uint32(data[4:8])
	result.Value = binary.LittleEndian.Uint32(data[8:12])

	return result
}

func loopIfds(data []byte, order uint16, offset int64, level int) IFDs {
	ifdLength := int64(12)

	var result IFDs
	var items []IFD

	result.Offset = offset

	entries, start := common.ReadUint16Order(data, order, offset)
	result.EntriesNr = entries

	for i := 0; i < int(entries); i++ {
		ifdbytes := data[int(start):int(start+ifdLength)]
		ifd := readIfd(ifdbytes)
		ifd.Level = level

		switch ifd.Tag {
		case 0x8769:
			//log.Printf("trovata EXIF subdirectory")
			ifd.SubIFDs = loopIfds(data, order, int64(ifd.Value), level+1)

		case 0x927c:
			log.Printf("trovata maker notes")
			ifd.SubIFDs = loopIfds(data, order, int64(ifd.Value), level+1)
		}
		/*
			if v, ok := Tags[ifd.Tag]; ok {
				//do something here
				log.Printf("Entry %s #%v", v, ifd)
			} else {
				log.Printf("Entry #%v", ifd)
			}
		*/
		items = append(items, ifd)
		start = start + ifdLength
	}

	var nextIfdOffset uint32
	nextIfdOffset, _ = common.ReadUint32Order(data, order, start)
	// log.Printf("nextIfdOffset=%d", nextIfdOffset)
	result.Ifds = items
	result.NextIfdOffset = int64(nextIfdOffset)
	return result
}
func readIfds(data []byte, header *Header) []IFDs {

	var result []IFDs
	var ifds IFDs
	var nextIfdOffset = header.IfdOffset

	for nextIfdOffset > 0 {
		ifds = loopIfds(data, header.ByteOrder, nextIfdOffset, 0)
		result = append(result, ifds)
		nextIfdOffset = ifds.NextIfdOffset
		//log.Printf("ifds:%v, nextOffset=%d", ifds, nextIfdOffset)
	}
	return result
}

func nSpaces(spaces int) string {
	var buffer bytes.Buffer

	for i := 0; i < spaces; i++ {
		buffer.WriteString("    ")
	}
	return buffer.String()

}
func dumpIfd(ifd IFD) {
	var desc string
	if v, ok := Tags[ifd.Level][ifd.Tag]; ok {
		desc = v
	} else {
		desc = "Tag "
	}
	log.Printf("%s %s #%v, Value=%v, Count=%d", nSpaces(ifd.Level), desc, ifd.Tag, ifd.Value, ifd.Count)
	for j := 0; j < len(ifd.SubIFDs.Ifds); j++ {
		ifd2 := ifd.SubIFDs.Ifds[j]
		dumpIfd(ifd2)
	}
}
func dumpIfds(ifds []IFDs) {
	for i := 0; i < len(ifds); i++ {
		log.Printf("IFD #%d\n", i)
		ifdrow := ifds[i]
		for k := 0; k < len(ifdrow.Ifds); k++ {
			ifd := ifdrow.Ifds[k]
			dumpIfd(ifd)

		}
	}
}
func Process(data []byte) {
	canonHeader, err := readHeader(data)
	check(err)
	log.Printf("Header %v\n", canonHeader)
	ifds := readIfds(data, &canonHeader)

	dumpIfds(ifds)
}
