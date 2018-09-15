package canon

import (
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReverseBitsIfNecessary(t *testing.T) {
	assert := assert.New(t)
	v1 := uint64(0x3f)
	v2 := reverseBitsIfNecessary(v1, 13)
	fmt.Printf("%b  -> %b\n", v1, v2)
	assert.Equal(uint64(0x1fc0), v2, "2")
}

func TestScanFile(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	dataCorrect, err := ioutil.ReadFile("../images/IMG_2060/IMG_2026_RAW_Data_from_DNG.bin")
	if err != nil {
		log.Printf("errore lettura file %v", err)
	}
	dataMaybe, err2 := ioutil.ReadFile("../ifd_3.bin")
	if err2 != nil {
		log.Printf("errore lettura file")
	}

	assert.Equal(len(dataCorrect), len(dataMaybe), "dimensione dei file non corrispondono")
	for i, b := range dataMaybe {
		require.Equal(b, dataCorrect[i], fmt.Sprintf("errore in posizione %d", i))
	}
}
