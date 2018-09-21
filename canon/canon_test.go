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

	dataCorrect, err := ioutil.ReadFile("../images/Canon/Canon_001_grayscale_lossless_jpeg.bin")
	if err != nil {
		log.Printf("errore lettura file %v", err)
	}
	dataMaybe, err2 := ioutil.ReadFile("../images/Canon/Canon_001.bin")
	if err2 != nil {
		log.Printf("errore lettura file")
	}

	assert.Equal(len(dataCorrect), len(dataMaybe), "dimensione dei file non corrispondono")
	for i, b := range dataMaybe {
		require.Equal(dataCorrect[i], b, fmt.Sprintf("errore in offset %d", i))
	}
}

func TestSliceIndex(t *testing.T) {
	assert := assert.New(t)

	slices := rawSlice{2, 1728, 1888}
	height := 3516

	v, rowInSlice, colInSlice := sliceIndex(0, slices, height)
	assert.Equal(0, v, "")
	assert.Equal(0, colInSlice, "")

	v, _, colInSlice = sliceIndex(1, slices, height)
	assert.Equal(0, v, "")
	assert.Equal(1, colInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(2*int(slices.SliceSize), slices, height)
	assert.Equal(0, v, "")
	assert.Equal(2, rowInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(2*int(slices.SliceSize)-1, slices, height)
	assert.Equal(0, v, "")
	assert.Equal(1, rowInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(6075647, slices, height)
	assert.Equal(0, v, "")

	v, rowInSlice, colInSlice = sliceIndex(6075648, slices, height)
	assert.Equal(1, v, "")
	assert.Equal(0, rowInSlice, "")
	assert.Equal(0, colInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(6075648+2*int(slices.SliceSize), slices, height)
	assert.Equal(1, v, "")
	assert.Equal(2, rowInSlice, "sono sulla riga 2 dello slice 2")

	// SLICE #2
	v, rowInSlice, colInSlice = sliceIndex(2*6075648, slices, height)
	assert.Equal(2, v, "")
	assert.Equal(0, rowInSlice, "")
	assert.Equal(0, colInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(2*6075648+1888, slices, height)
	assert.Equal(2, v, "")
	assert.Equal(1, rowInSlice, "")
	assert.Equal(0, colInSlice, "")

	v, rowInSlice, colInSlice = sliceIndex(5344*3516-1, slices, height)
	assert.Equal(2, v, "")
	assert.Equal(3515, rowInSlice, "")
	assert.Equal(1887, colInSlice, "")

}

func TestTypeConv(t *testing.T) {
	assert := assert.New(t)

	var a int32
	var b int32

	a = 10
	b = -20

	c := a + b
	d := uint64(c)
	assert.Equal(fmt.Sprintf("%b", c), fmt.Sprintf("%b", d), "")
}
