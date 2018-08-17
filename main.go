package main

import (
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

// identify identifies the file maker
func identify(inputFile *os.File) inputImage {

	result := inputImage{make: "UNDEF"}

	head := make([]byte, 32)
	n1, err := inputFile.Read(head)
	check(err)
	fmt.Printf("%d bytes : %s\n", n1, string(head))

	if strings.HasPrefix(string(head), "FUJIFILM") {
		result.make = "FUJIFILM"
	}

	return result

}
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	rawfile := flag.String("f", "", "raw file")
	flag.Parse()
	log.Println("reading file " + *rawfile)

	inputFile, err := os.Open(*rawfile)
	check(err)

	inputImage := identify(inputFile)

	fmt.Printf("Make: %s\n", inputImage.make)
	//
}
