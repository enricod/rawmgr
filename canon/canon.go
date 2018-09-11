package canon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"

	bitstream "github.com/dgryski/go-bitstream"
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

// DHTHeader -
type DHTHeader struct {
	Marker      uint16
	Length      uint16
	TableClass0 uint8
	TableIndex0 uint8
	TableClass1 uint8
	TableIndex1 uint8
}

type LosslessJPG struct {
	DHTHeader    DHTHeader
	SOF3Header   SOF3Header
	SOSHeader    SOSHeader
	HuffmanCodes [][]common.HuffMapping
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

func dumpIfd(ifd IFD) {
	var desc string
	if v, ok := Tags[ifd.Level][ifd.Tag]; ok {
		desc = v
	} else {
		desc = "Tag "
	}
	log.Printf("%s %s #%v, Value=%v, Count=%d", common.NSpaces(ifd.Level), desc, ifd.Tag, ifd.Value, ifd.Count)
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

/*
ushort * CLASS make_decoder_ref (const uchar **source)
{
    int max, len, h, i, j;
    const uchar *count;
    ushort *huff;

    count = (*source += 16) - 17;
    for (max=16; max && !count[max]; max--);
    huff = (ushort *) calloc (1 + (1 << max), sizeof *huff);
    merror (huff, "make_decoder()");
    huff[0] = max;
    for (h=len=1; len <= max; len++)
        for (i=0; i < count[len]; i++, ++*source)
            for (j=0; j < 1 << (max-len); j++)
                if (h <= 1 << max)
                    huff[h++] = len << 8 | **source;
    return huff;
}
*/
/*
   Construct a decode tree according the specification in *source.
   The first 16 bytes specify how many codes should be 1-bit, 2-bit
   3-bit, etc.  Bytes after that are the leaf values.

   For example, if the source is

    { 0,1,4,2,3,1,2,0,0,0,0,0,0,0,0,0,
      0x04,0x03,0x05,0x06,0x02,0x07,0x01,0x08,0x09,0x00,0x0a,0x0b,0xff  },

   then the code is

	00		0x04
	010		0x03
	011		0x05
	100		0x06
	101		0x02
	1100		0x07
	1101		0x01
	11100		0x08
	11101		0x09
	11110		0x00
	111110		0x0a
	1111110		0x0b
	1111111		0xff
*/

type huffItem struct {
	Key  []byte
	Code uint16
}

type SOF3Component struct {
	ComponentIdentifier      uint8
	HorizontalSamplingFactor uint8
	VerticalSamplingFactor   uint8
	QuantizationTable        uint8
}

// SOF3Header - start of frame header
type SOF3Header struct {
	Marker                    uint16
	Length                    uint16
	SamplePrecision           uint8
	NrLines                   uint16
	NrSamplesPerLine          uint16
	NrImageComponentsPerFrame uint8
	Components                []SOF3Component
}

type SOSComponent struct {
	Selector uint8
	DCTable  uint8
	ACTable  uint8
}

// SOSHeader - start of scan header
type SOSHeader struct {
	Marker                              uint16
	Length                              uint16
	NrComponents                        uint8
	Components                          []SOSComponent
	StartOfSpectral                     uint8
	EndOfSpectral                       uint8
	SuccessiveApprosimationBitPositions uint8
}

func hammingDistance(a, b []byte) (int, error) {
	if len(a) != len(b) {
		return 0, errors.New("a b are not the same length")
	}

	diff := 0
	for i := 0; i < len(a); i++ {
		b1 := a[i]
		b2 := b[i]
		for j := 0; j < 8; j++ {
			mask := byte(1 << uint(j))
			if (b1 & mask) != (b2 & mask) {
				diff++
			}
		}
	}
	return diff, nil
}

// SOF3 start of frame
func parseSOF3Header(data []byte, offset int64) (SOF3Header, int64, error) {
	//log.Printf("SOF3 header offset %d", offset)

	sof3Header := SOF3Header{}
	marker, offset2 := common.ReadUint16(data, offset)

	log.Printf("SOF3 offset=%d, marker=%d", offset, marker)
	if marker != 0xffc3 {
		_, err := fmt.Printf("SOF3 header invalid, expected %d, found %d", 0xffc3, marker)
		return sof3Header, offset2, err
	}
	sof3Header.Marker = marker
	length, offset2 := common.ReadUint16(data, offset2)
	sof3Header.Length = length

	samplePrecision, offset2 := common.ReadUint8(data, offset2)
	sof3Header.SamplePrecision = samplePrecision

	log.Printf("SOF3 samplePrecision=%d", samplePrecision)

	nrLines, offset2 := common.ReadUint16(data, offset2)
	sof3Header.NrLines = nrLines

	nrSamplePerLine, offset2 := common.ReadUint16(data, offset2)
	sof3Header.NrSamplesPerLine = nrSamplePerLine

	imageComponentsPerFrame, offset2 := common.ReadUint8(data, offset2)
	sof3Header.NrImageComponentsPerFrame = imageComponentsPerFrame

	log.Printf("nrSamplePerLine=%d, imageComponentsPerFrame=%d ", nrSamplePerLine, imageComponentsPerFrame)
	// let's read each component
	components := []SOF3Component{}
	var offset3 = offset2
	var identifier, quantizationTable uint8

	for i := 0; i < int(imageComponentsPerFrame); i++ {
		identifier, offset3 = common.ReadUint8(data, offset3)
		samplingByte := data[offset3]
		offset3++
		quantizationTable, offset3 = common.ReadUint8(data, offset3)

		comp := SOF3Component{ComponentIdentifier: identifier,
			HorizontalSamplingFactor: uint8(samplingByte >> 4),
			VerticalSamplingFactor:   uint8(samplingByte & 0x0f),
			QuantizationTable:        quantizationTable}
		components = append(components, comp)
	}

	sof3Header.Components = components

	return sof3Header, offset3, nil
}

func parseSOSHeader(data []byte, offset int64) (SOSHeader, int64, error) {
	sosHeader := SOSHeader{}
	marker, offset2 := common.ReadUint16(data, offset)
	log.Printf("SOS  offset=%d, marker=%d", offset, marker)
	if marker != 0xffda {
		_, err := fmt.Printf("SOS header invalid, expected %d, found %d", 0xffda, marker)
		return sosHeader, offset2, err
	}
	sosHeader.Marker = marker

	length, offset2 := common.ReadUint16(data, offset2)
	sosHeader.Length = length

	nrComponents, offset2 := common.ReadUint8(data, offset2)
	sosHeader.NrComponents = nrComponents
	log.Printf("SOS  nrComponents=%d", nrComponents)

	// let's read each component
	components := []SOSComponent{}
	var offset3 = offset2
	var identifier uint8

	for i := 0; i < int(nrComponents); i++ {
		identifier, offset3 = common.ReadUint8(data, offset3)
		samplingByte := data[offset3]
		offset3++

		comp := SOSComponent{Selector: identifier,
			DCTable: uint8(samplingByte >> 4),
			ACTable: uint8(samplingByte & 0x0f),
		}

		log.Printf("SOS  component # %d, table=%d", comp.Selector, comp.DCTable)
		components = append(components, comp)
	}
	sosHeader.Components = components

	startOfSpectral, offset3 := common.ReadUint8(data, offset3)
	sosHeader.StartOfSpectral = startOfSpectral

	endOfSpectral, offset3 := common.ReadUint8(data, offset3)
	sosHeader.EndOfSpectral = endOfSpectral

	successiveApprosimationBitPositions, offset3 := common.ReadUint8(data, offset3)
	sosHeader.SuccessiveApprosimationBitPositions = successiveApprosimationBitPositions

	return sosHeader, offset3, nil
}

func parseDHTHeader(data []byte, offset int64) (LosslessJPG, int64, error) {
	var dhtHeader = DHTHeader{}

	log.Printf("parseDHTHeader, offset=%d\n", offset)
	marker, offset2 := common.ReadUint16(data, offset)

	if marker != 0xffc4 {
		return LosslessJPG{}, offset2, fmt.Errorf("DHT Marker not valid  %d", marker)
	}

	dhtHeader.Marker = marker

	// log.Printf("length offset %d", offset2)
	length, offset2 := common.ReadUint16(data, offset2)
	dhtHeader.Length = length

	huffBytes := data[offset : offset+int64(length-2)]
	huffMappings := common.DecodeHuffTree(huffBytes)

	sof3Header, offset2, err := parseSOF3Header(data, offset2+int64(dhtHeader.Length)-2)
	if err != nil {
		return LosslessJPG{}, offset2, err
	}

	sosHeader, offset3, err := parseSOSHeader(data, offset2)
	if err != nil {
		return LosslessJPG{}, offset3, err
	}

	losslessJPG := LosslessJPG{DHTHeader: dhtHeader, SOF3Header: sof3Header,
		SOSHeader:    sosHeader,
		HuffmanCodes: huffMappings,
	}

	return losslessJPG, offset3, nil
}

func getRawSlice(ifd IFDs) (rawSlice, error) {
	for _, ifd := range ifd.Ifds {
		if ifd.Tag == 0xc640 {
			return ifd.RawSlice, nil
		}
	}
	return rawSlice{}, errors.New("raw slice not found")
}

// if  0xff00 is followed by 0x00 we return only 0xff - deprecated
func extractFirstBytes(data []byte, offset int64, howmany int) ([]byte, int64) {
	mybytes := []byte{}

	var pos = offset
	i := 0
	for len(mybytes) < howmany {
		b := data[pos+int64(i)]
		mybytes = append(mybytes, b)
		if b == 0xff && data[pos+int64(i)+1] == 0x00 {
			i++
		}
		i++
	}
	return mybytes, offset + int64(i)
}

func findHuffMapping(mappings []common.HuffMapping, mycode uint64) (common.HuffMapping, error) {
	if mycode == uint64(0) {
		return common.HuffMapping{}, fmt.Errorf("not found")
	}

	myvalue, err := common.HuffGetMapping(mappings, mycode)

	if err != nil {
		return findHuffMapping(mappings, mycode>>1)
	} else {
		return myvalue, nil
	}
}

// if first bit == 0, then do reverse
func reverseBitsIfNecessary(a uint64, bitNr int) uint64 {
	p2 := common.Pow2(bitNr)
	mask := p2 >> 1 // something like 100...00
	if a&mask == 0 {
		// if first bit == 0
		return uint64(p2-1) - a
	}
	return a
}

func findHuffCode(data []byte, bitsOffset int, bitsLength int, huffMappings []common.HuffMapping) (common.HuffMapping, error) {
	bitreader := bitstream.NewReader(bytes.NewReader(data))
	bitreader.ReadBits(bitsOffset)
	v, err := bitreader.ReadBits(bitsLength)
	if err != nil {
		return common.HuffMapping{}, err
	}

	h, err2 := common.HuffGetMapping(huffMappings, v)
	if err2 != nil {
		return findHuffCode(data, bitsOffset, bitsLength-1, huffMappings)
	}
	return h, nil
}

// 0xff 0x00 becomes 0xff
func cleanStream(data []byte) []byte {
	result := []byte{}
	for i, b := range data {
		if !(i > 0 && b == 0x00 && data[i-1] == 0xff) {
			result = append(result, b)
		}
	}
	return result
}

func scanRawData(data []byte, loselessJPG LosslessJPG, offset int64, canonHeader Header, aifd IFDs) error {

	cleanedData := cleanStream(data[offset:])
	log.Printf("size %d, cleaned %d, removed %d", len(data[offset:]), len(cleanedData), len(data[offset:])-len(cleanedData))

	initialValue := common.Pow2(int(loselessJPG.SOF3Header.SamplePrecision - 1))

	log.Printf("scanRawData | offset=%d, initial value %d", offset, initialValue)
	rawSlice, err := getRawSlice(aifd)
	if err != nil {
		return err
	}
	log.Printf("rawSlice %v", rawSlice)

	componentNr := 0

	rawData := []uint64{}
	bitsOffset := 0

	bitreader := bitstream.NewReader(bytes.NewReader(cleanedData))

	// PROVVISORIO
	for j := 0; j < int(loselessJPG.SOF3Header.NrImageComponentsPerFrame); j++ {
		log.Printf("============= STEP %d ===============", j)
		log.Printf("bitsOffset = %d", bitsOffset)
		dcTableIndex := int(loselessJPG.SOSHeader.Components[componentNr].DCTable)
		log.Printf("componentNr=%d, dcTableIndex = %d", componentNr, dcTableIndex)

		huffCode, err := findHuffCode(cleanedData, bitsOffset, 16, loselessJPG.HuffmanCodes[dcTableIndex])
		if err != nil {
			log.Printf(err.Error())
		}
		log.Printf("%v", huffCode)

		// we already read huffCode.BitCount bits searching huffCode
		bitreader.ReadBits(huffCode.BitCount)
		val2, err := bitreader.ReadBits(int(huffCode.Value))

		log.Printf("val2=%d, %13b", val2, val2)

		if err != nil {
			log.Printf(err.Error())
		}

		val3 := reverseBitsIfNecessary(val2, int(huffCode.Value))
		log.Printf("val2=%13b, val3=%13b", val2, val3)
		var val4 uint64
		if val3 == val2 {
			// highest bit = 1
			val4 = initialValue + val3
		} else {
			// highest bit = 0
			val4 = initialValue - val3
		}

		// costruiamo byte per immagine finale
		/*
			val5Bytes := make([]byte, 2)
			binary.LittleEndian.PutUint16(val5Bytes, uint16(val4))
			log.Printf("val4 = %d,  %v", val4, val5Bytes)

			rawData = append(rawData, val5Bytes...)
		*/
		rawData = append(rawData, val4)
		log.Printf(" %v  ", rawData)

		// prepare for next iteration
		componentNr++
		if componentNr > int(loselessJPG.SOF3Header.NrImageComponentsPerFrame) {
			componentNr = 0
		}
		bitsOffset = bitsOffset + huffCode.BitCount + int(huffCode.Value)
	}
	return nil
}

func parseRaw(data []byte, canonHeader Header, aifd IFDs, filename string) error {
	startOffset, _ := getStartEndIFD0(aifd)

	soiMarker, offset := common.ReadUint16(data, startOffset)
	if soiMarker != 0xffd8 {
		return fmt.Errorf("SOI Marker not valid  %d", soiMarker)
	}

	loselessJPG, loselessJPGOffset, err := parseDHTHeader(data, offset)
	check(err)
	log.Printf("loselessJPG %v", loselessJPG)

	scanRawData(data, loselessJPG, loselessJPGOffset, canonHeader, aifd)

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
