package canon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReverseBitsIfNecessary(t *testing.T) {
	assert := assert.New(t)
	v1 := uint64(0x3f)
	v2 := reverseBitsIfNecessary(v1, 13)
	fmt.Printf("%b  -> %b\n", v1, v2)
	assert.Equal(uint64(0x1fc0), v2, "2")

}
