package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"os"

	"github.com/enricod/rawmgr/common"
)

// Tags name
var Tags = []map[uint16]string{
	{
		0x0100: "width",
		0x0101: "height",
		272:    "model",
		0x0111: "stripOffset",
		0x0112: "orientation",
		0x0117: "stripByteCounts",
		0x011a: "xResolution",
		0x8729: "exif",
		0xc640: "cr2Slice",
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

// Header for canon file
type Header struct {
	ByteOrder       uint16
	TiffMagicWord   uint16
	IfdOffset       int64
	CR2MagicWord    string
	CR2MajorVersion uint8
	CR2MinorVersion uint8
	RawIfdOffset    int64
}

type rawSlice struct {
	Count         uint16
	SliceSize     uint16
	LastSliceSize uint16
}

// IFD Image File Directory item
type IFD struct {
	Tag           uint16
	Typ           uint16
	Count         uint32
	Value         uint32
	ValueAsString string
	Level         int
	SubIFDs       IFDs
	RawSlice      rawSlice
}

// IFDs Image File Directory
type IFDs struct {
	EntriesNr     uint16
	Offset        int64
	Ifds          []IFD
	NextIfdOffset int64
}

type DHTHeader struct {
	Marker     uint16
	Length     uint16
	TableClass uint8
	TableIndex uint8
}

type SOF3Header struct {
}

type SOSHeader struct {
}

func readHeader(data []byte) (Header, error) {
	var start int64
	var result = Header{}

	// "II" or 0x4949 (18761) means Intel byte order (little endian)
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
			// EXIF subdirectory
			ifd.SubIFDs = loopIfds(data, order, int64(ifd.Value), level+1)

		case 0x927c:
			// maker notes
			ifd.SubIFDs = loopIfds(data, order, int64(ifd.Value), level+1)

		case 0xC640:
			// SLICES
			var sliceCount, sliceSize, lastSliceSize uint16
			var nextOffset int64
			sliceCount, nextOffset = common.ReadUint16Order(data, order, int64(ifd.Value))
			sliceSize, nextOffset = common.ReadUint16Order(data, order, nextOffset)
			lastSliceSize, _ = common.ReadUint16Order(data, order, nextOffset)
			var aRawSlice = rawSlice{Count: sliceCount, SliceSize: sliceSize, LastSliceSize: lastSliceSize}
			ifd.RawSlice = aRawSlice
			log.Printf("Slice %v \n", aRawSlice)

		}

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

type calcStartEnd func(ifd IFDs) (int64, int64)

var getStartEndIFD0 = calcStartEnd(func(aifd IFDs) (int64, int64) {
	var start, end, bytesCount int64
	for j := 0; j < len(aifd.Ifds); j++ {
		ifd := aifd.Ifds[j]
		switch ifd.Tag {
		case 273:
			start = int64(ifd.Value)
		case 279:
			bytesCount = int64(ifd.Value)

		}
	}
	end = start + bytesCount
	return start, end
})

var getStartEndIFD1 = calcStartEnd(func(aifd IFDs) (int64, int64) {
	var start, end, bytesCount int64
	for j := 0; j < len(aifd.Ifds); j++ {
		ifd := aifd.Ifds[j]
		switch ifd.Tag {
		case 0x201:
			start = int64(ifd.Value)
		case 0x202:
			bytesCount = int64(ifd.Value)
		}
	}
	end = start + bytesCount
	return start, end
})

func saveJpeg(data []byte, aifd IFDs, filename string, calc calcStartEnd) {
	start, end := calc(aifd)
	log.Printf("Saving JPEG %d -> %d", start, end)
	jpegData := data[start:end]
	f, err := os.Create(filename)
	_, err = f.Write(jpegData)
	check(err)
	defer f.Close()

}

func parseDHTHeader(data []byte, offset int64) (DHTHeader, error) {
	var dhtHeader = DHTHeader{}

	marker, offset2 := common.ReadUint16(data, offset)
	if marker != 0xffc4 {
		return dhtHeader, fmt.Errorf("DHT Marker not valid  %d", marker)
	}

	dhtHeader.Marker = marker

	length, offset2 := common.ReadUint16(data, offset2)
	dhtHeader.Length = length

	tableClass := uint8(data[offset2])
	dhtHeader.TableClass = tableClass
	offset2++

	tableIndex := uint8(data[offset2])
	dhtHeader.TableIndex = tableIndex
	offset2++

	return dhtHeader, nil
}

func parseRaw(data []byte, canonHeader Header, aifd IFDs, filename string) error {
	startOffset, _ := getStartEndIFD0(aifd)

	soiMarker, offset := common.ReadUint16(data, startOffset)
	if soiMarker != 0xffd8 {
		return fmt.Errorf("SOI Marker not valid  %d", soiMarker)
	}

	dhtHeader, err := parseDHTHeader(data, offset)
	check(err)
	log.Printf("DHTHeader %v", dhtHeader)
	return nil
}

// ProcessCR2 start CR2 files
func ProcessCR2(data []byte) {
	canonHeader, err := readHeader(data)
	check(err)
	log.Printf("Header %v\n", canonHeader)
	ifds := readIfds(data, &canonHeader)

	if *common.Verbose {
		dumpIfds(ifds)
	}
	saveJpeg(data, ifds[0], "ifd_0.jpeg", getStartEndIFD0)
	saveJpeg(data, ifds[1], "ifd_1.jpeg", getStartEndIFD1)

	err = parseRaw(data, canonHeader, ifds[3], "ifd_3.jpeg")
	check(err)
}
