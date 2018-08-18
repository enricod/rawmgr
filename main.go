package main

import (
	"encoding/binary"
	"flag"
	"log"
	"os"
	"strings"
)

type imageInfo struct {
	make string
}

const RawmgrVersion = "0.1"

func xtransInterpolate() {

}

// getUint16 a partire da offset legge un int uint16 e torna la nuova posizione
func getUint16(f *os.File, offset int64) (uint16, int64) {
	value := binary.BigEndian.Uint16(readFromFileBytes(f, offset, 2))
	return value, offset + 2
}

func get1Byte(f *os.File, offset int64) (uint16, int64) {
	mybyte := readFromFileBytes(f, offset, 1)[0]
	return uint16(mybyte), offset + 1
}

func readFromFile(f *os.File, offset int64, howmany int64) (uint32, int64) {
	value := binary.BigEndian.Uint32(readFromFileBytes(f, offset, howmany))
	return value, offset + howmany
}
func parseFuji(f *os.File, offset int64) {

	var tag, len uint16
	var start, posizione int64
	var rawHeight, rawWidth uint16
	var height, fujiLayout uint16
	//var width uint32
	//var fujiWidth uint32
	var filters int
	var xtransAbs [6][6]uint16

	start = offset
	entries, start := readFromFile(f, start, 4)

	log.Printf("entries %d \n", entries)
	if entries < 255 {

		for i := 0; i < int(entries); i++ {
			/*
				tag = get2();
				len = get2();
				save = ftell(ifp);
			*/
			tag, start = getUint16(f, start)
			len, start = getUint16(f, start)
			posizione = start
			log.Printf("posizione=%d, tag=%d, len=%d", posizione, tag, len)
			switch tag {
			case 0x100:
				rawHeight, start = getUint16(f, start)
				rawWidth, start = getUint16(f, start)
				log.Printf("raw_width=%d, raw_height=%d", rawHeight, rawWidth)

			case 0x121:
				height, start = getUint16(f, start)
				log.Printf("height=%d", height)

			case 0x130:
				fujiLayout, start = getUint16(f, start)
				fujiLayout = fujiLayout >> 7
				log.Printf("fujiLayout=%d", fujiLayout)
			//fujiWidth = !(fgetc(ifp) & 8)
			case 0x131:
				filters = 9
				var val uint16
				for r := 5; r >= 0; r-- {
					for c := 5; c >= 0; c-- {
						val, start = get1Byte(f, start)
						xtransAbs[r][c] = val & 3
					}
				}
				log.Printf("filters=%d, xtransAbs=%v", filters, xtransAbs)
				/*
					filters = 9;
					FORC(36) xtrans_abs[0][35-c] = fgetc(ifp) & 3;
				*/

			case 0x2ff0:
				log.Printf("WARN tag non ancora elaborato=%d", tag)
			case 0xc000:
				if len > 20000 {
					/*
						c = order;
						order = 0x4949;
						while ((tag = get4()) > raw_width);
						width = tag;
						height = get4();
						order = c;
					*/

				}
			default:

			}

			start = posizione + int64(len)
		}

	}

}

func fujiXtransRead(inputFile *os.File) {

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

// identify identifies the file maker
func identify(inputFile *os.File) imageInfo {
	result := imageInfo{make: "UNDEF"}
	head := make([]byte, 32)
	_, err := inputFile.Read(head)
	check(err)
	//fmt.Printf("%d bytes : %s\n", n1, string(head))

	if strings.HasPrefix(string(head), "FUJIFILM") {
		result.make = "FUJIFILM"
		thumbOffset := binary.BigEndian.Uint32(readFromFileBytes(inputFile, 84, 4))
		thumbLength := binary.BigEndian.Uint32(readFromFileBytes(inputFile, 88, 4))
		log.Printf("thumbOffset=%d, thumbLength=%d\n", thumbOffset, thumbLength)

		startParse, _ := readFromFile(inputFile, 92, 4)
		parseFuji(inputFile, int64(startParse))
	}

	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	rawfile := flag.String("f", "DSCF2483.RAF", "raw file")
	flag.Parse()
	log.Println("reading file " + *rawfile)

	inputFile, err := os.Open(*rawfile)
	check(err)

	imageInfo := identify(inputFile)

	log.Printf("Make: %s\n", imageInfo.make)
	//
}
