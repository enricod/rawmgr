package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
)

const RawmgrVersion = "0.1"

func xtransInterpolate() {

}
func fujiXtransRead(bytes []byte) {
	fmt.Printf("Hello, world.\n")
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

	dat, err := ioutil.ReadFile(*rawfile)
	check(err)

	fujiXtransRead(dat)
	//
}
