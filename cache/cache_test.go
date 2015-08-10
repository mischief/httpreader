package httpreader

import (
	"bytes"
	"testing"
)

const (
	blockSize = 512
	maxBlocks = 100
)

func TestCache(t *testing.T) {
	c := NewCache(blockSize, maxBlocks, maxBlocks*blockSize)

	for i := 0; i < maxBlocks+1; i++ {
		blk := &cacheBlock{
			offset: int64(i * blockSize),
			data:   bytes.Repeat([]byte{byte(i)}, blockSize),
		}
		c.addBlock(blk)

		// check for presence
		cblk := c.getBlock(0, false)
		if cblk == nil {
			t.Errorf("%d: expected cache block 0 got nil", i)
			return
		}
	}

	// check for eviction
	cblk := c.getBlock(blockSize, false)
	if cblk != nil {
		t.Errorf("expected nil got %v", cblk)
		return
	}
}
