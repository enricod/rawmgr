package main

import (
	"flag"

	"log"
	"os"
	"strings"

	"github.com/enricod/rawmgr/common"
	"github.com/enricod/rawmgr/fuji"
)

type imageInfo struct {
	make string
}

func xtransInterpolate(passes int) {
	fuji.XtransInterpolate(passes)
}

const RawmgrVersion = "0.1"

func fujiXtransRead(inputFile *os.File) {

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
		var thumbOffset, thumbLength uint32
		var start int64
		start = 84
		thumbOffset, start = common.GetUint32(inputFile, start)
		thumbLength, start = common.GetUint32(inputFile, start)
		log.Printf("thumbOffset=%d, thumbLength=%d\n", thumbOffset, thumbLength)

		startParse, _ := common.GetUint32(inputFile, 92)
		fuji.ParseFuji(inputFile, int64(startParse))

		dataOffset, _ := common.GetUint32(inputFile, 100)
		common.ParseTiff(inputFile, int64(dataOffset))

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
