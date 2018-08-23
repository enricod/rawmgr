package main

import (
	"flag"
	"fmt"

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

	var start int64
	var hlen uint32
	var order uint16

	order, start = common.GetUint16(inputFile, 0)
	hlen, start = common.GetUint32(inputFile, start)

	fmt.Printf("order=%d, hlen=%d, start=%d\n", order, hlen, start)
	result := imageInfo{make: "UNDEF"}
	head := make([]byte, 32)
	_, err := inputFile.ReadAt(head, 0)
	check(err)
	//fmt.Printf("%d bytes : %s\n", n1, string(head))
	var tiffIfsArray []common.TiffIfd
	tiffIfsArray = make([]common.TiffIfd, 0)

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

		common.ParseTiff(inputFile, int64(dataOffset), tiffIfsArray)

	} else if order == 0x4949 || order == 0x4d4d {
		// Canon?
		tiffIfdArray2 := common.ParseTiff(inputFile, 0, tiffIfsArray)
		fmt.Printf("tiffIfdArray2=%d\n", len(tiffIfdArray2))
	}

	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	defaultFileName := "IMG_2509.CR2" // "DSCF2483.RAF" // "IMG_2509.CR2"
	rawfile := flag.String("f", defaultFileName, "raw file")
	flag.Parse()
	log.Println("reading file " + *rawfile)

	inputFile, err := os.Open(*rawfile)
	check(err)

	imageInfo := identify(inputFile)

	log.Printf("Make: %s\n", imageInfo.make)
	//
}
