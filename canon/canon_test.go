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
	dataMaybe, err2 := ioutil.ReadFile("../ifd_3.bin")
	if err2 != nil {
		log.Printf("errore lettura file")
	}

	assert.Equal(len(dataCorrect), len(dataMaybe), "dimensione dei file non corrispondono")
	for i, b := range dataMaybe {
		require.Equal(dataCorrect[i], b, fmt.Sprintf("errore in posizione %d", i))
	}
}

func TestSliceIndex(t *testing.T) {
	assert := assert.New(t)

	slices := rawSlice{2, 1728, 1888}
	height := 3516
	v, rowInSlice := sliceIndex(0, slices, height)
	assert.Equal(0, v, "")

	v, _ = sliceIndex(1, slices, height)
	assert.Equal(0, v, "")

	v, rowInSlice = sliceIndex(2*int(slices.SliceSize), slices, height)
	assert.Equal(0, v, "")
	assert.Equal(2, rowInSlice, "")
	v, rowInSlice = sliceIndex(2*int(slices.SliceSize)-1, slices, height)
	assert.Equal(0, v, "")
	assert.Equal(1, rowInSlice, "")

	v, _ = sliceIndex(6075647, slices, height)
	assert.Equal(0, v, "")

	v, rowInSlice = sliceIndex(6075648+2*int(slices.SliceSize), slices, height)
	assert.Equal(1, v, "")
	assert.Equal(2, rowInSlice, "sono sulla riga 2 dello slice 2")

	v, rowInSlice = sliceIndex(2*6075648, slices, height)
	assert.Equal(2, v, "")

	v, rowInSlice = sliceIndex(2*6075648+3*int(slices.LastSliceSize), slices, height)
	assert.Equal(2, v, "")
	assert.Equal(3, rowInSlice, "sono sulla riga 3 dello slice 3")
}

/*
func TestGetPositionWithoutSlicing(t *testing.T) {
	assert := assert.New(t)

	slices := rawSlice{2, 1728, 1888}
	nrLines := 3516

	// primo elemento, primo slice
	sliceIndex, sliceRow, sliceCol, index := getPositionWithoutSlicing(0, slices, nrLines)
	assert.Equal(0, sliceIndex, "")
	assert.Equal(0, sliceRow, "")
	assert.Equal(0, sliceCol, "")
	assert.Equal(0, index, "")

	// slice #0, row #1, col 0
	sliceIndex, sliceRow, sliceCol, _ = getPositionWithoutSlicing(1728, slices, nrLines)
	assert.Equal(0, sliceIndex, "")
	assert.Equal(1, sliceRow, "")
	assert.Equal(0, sliceCol, "")

	// slice #0, row #1, col 0
	sliceIndex, sliceRow, sliceCol, _ = getPositionWithoutSlicing(1730, slices, nrLines)
	assert.Equal(0, sliceIndex, "")
	assert.Equal(1, sliceRow, "")
	assert.Equal(2, sliceCol, "")

	// slice #1, row #0, col 0
	sliceIndex, sliceRow, sliceCol, index = getPositionWithoutSlicing(6075648, slices, nrLines)
	assert.Equal(1, sliceIndex, "")
	assert.Equal(0, sliceRow, "")
	assert.Equal(0, sliceCol, "")
	assert.Equal(1728, index, "")

	sliceIndex, sliceRow, sliceCol, index = getPositionWithoutSlicing(6075648+1728, slices, nrLines)
	assert.Equal(1, sliceIndex, "")
	assert.Equal(1, sliceRow, "")
	assert.Equal(0, sliceCol, "")
	assert.Equal(5344+1728, index, "hh")

	sliceIndex, sliceRow, sliceCol, index = getPositionWithoutSlicing(6075648+2*1728+10, slices, nrLines)
	assert.Equal(1, sliceIndex, "")
	assert.Equal(2, sliceRow, "")
	assert.Equal(10, sliceCol, "")
	assert.Equal(5344*2+1728*sliceIndex+10, index, "")

	// slice #2
	sliceIndex, sliceRow, sliceCol, index = getPositionWithoutSlicing(2*6075648, slices, nrLines)
	assert.Equal(2, sliceIndex, "")
	assert.Equal(0, sliceRow, "")
	assert.Equal(0, sliceCol, "")
	assert.Equal(sliceIndex*1728, index, "")

	sliceIndex, sliceRow, sliceCol, index = getPositionWithoutSlicing(2*6075648+1728, slices, nrLines)
	assert.Equal(2, sliceIndex, "")
	assert.Equal(0, sliceRow, "")
	assert.Equal(1728, sliceCol, "")

	sliceIndex, sliceRow, sliceCol, _ = getPositionWithoutSlicing(2*6075648+1888, slices, nrLines)
	assert.Equal(2, sliceIndex, "")
	assert.Equal(1, sliceRow, "")
	assert.Equal(0, sliceCol, "")
}
*/
