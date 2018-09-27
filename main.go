package main

import (
	"flag"
	"fmt"
	"image/png"
	"io/ioutil"
	"runtime/pprof"

	"log"
	"os"
	"strings"

	"github.com/enricod/rawmgr/canon"
	"github.com/enricod/rawmgr/common"
	"github.com/enricod/rawmgr/fuji"
)

type imageInfo struct {
	make string
}

func xtransInterpolate(passes int) {
	fuji.XtransInterpolate(passes)
}

func fujiXtransRead(inputFile *os.File) {

}

// identify identifies the file maker
func identify(inputFile *os.File) imageInfo {

	var start int64
	var hlen uint32
	var order uint16

	order, start = common.GetUint16(inputFile, 0)
	hlen, start = common.GetUint32(inputFile, start)

	if *common.Verbose {
		fmt.Printf("order=%d, hlen=%d, start=%d\n", order, hlen, start)
	}
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
		if *common.Verbose {
			log.Printf("thumbOffset=%d, thumbLength=%d\n", thumbOffset, thumbLength)
		}
		startParse, _ := common.GetUint32(inputFile, 92)
		fuji.ParseFuji(inputFile, int64(startParse))

		dataOffset, _ := common.GetUint32(inputFile, 100)

		common.ParseTiff(inputFile, int64(dataOffset), tiffIfsArray)

	} else if order == 0x4949 || order == 0x4d4d {

		data, err := ioutil.ReadFile(inputFile.Name())
		check(err)
		canon.ProcessCR2(data, "PPP")

	}

	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	argsWithoutProg := os.Args[1:]
	log.Printf("%v", argsWithoutProg)
	if len(argsWithoutProg) == 0 {
		fmt.Printf("input file not specified \n")
		return
	}

	//defaultFileName := "images/Canon/Canon_001.CR2"
	rawfile := argsWithoutProg[len(argsWithoutProg)-1]
	common.Verbose = flag.Bool("v", false, "verbose")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	common.ShowInfo = flag.Bool("i", false, "show image info")
	common.ExtractJpegs = flag.Bool("j", false, "extract jpegs")

	flag.Parse()
	log.Println("reading file " + rawfile)

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	/*
		inputFile, err := os.Open(*rawfile)
		check(err)

		imageInfo := identify(inputFile)

		log.Printf("Make: %s\n", imageInfo.make)
	*/

	data, err := ioutil.ReadFile(rawfile)
	check(err)

	imageRGBA := canon.ProcessCR2(data, rawfile)

	outputFile, err := os.Create(strings.Replace(rawfile, ".CR2", ".png", 1))
	if err != nil {
		// Handle error
	}

	png.Encode(outputFile, imageRGBA)

	// Don't forget to close files
	outputFile.Close()
	log.Printf("wrote %s", outputFile.Name())
}
