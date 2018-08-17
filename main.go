package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type inputImage struct {
	make string
}

const RawmgrVersion = "0.1"

func xtransInterpolate() {

}
func fujiXtransRead(inputFile *os.File) {

}

func readFromFileBytes(f *os.File, start int64, howmany int64) []byte {
	o2, err := f.Seek(start, 0)
	check(err)
	thumbOffsetBytes := make([]byte, howmany)
	n2, err := f.Read(thumbOffsetBytes)
	check(err)
	fmt.Printf("%d bytes @ %d: %s\n", n2, o2, string(thumbOffsetBytes))
	return thumbOffsetBytes
}

// identify identifies the file maker
func identify(inputFile *os.File) inputImage {

	result := inputImage{make: "UNDEF"}

	head := make([]byte, 32)
	_, err := inputFile.Read(head)
	check(err)
	//fmt.Printf("%d bytes : %s\n", n1, string(head))

	if strings.HasPrefix(string(head), "FUJIFILM") {
		result.make = "FUJIFILM"

		thumbOffset := binary.BigEndian.Uint32(readFromFileBytes(inputFile, 84, 4))
		thumbLength := binary.BigEndian.Uint32(readFromFileBytes(inputFile, 88, 4))
		fmt.Println(thumbOffset)
		fmt.Println(thumbLength)
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

	inputImage := identify(inputFile)

	fmt.Printf("Make: %s\n", inputImage.make)
	//
}
