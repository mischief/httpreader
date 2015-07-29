package httpreader

import (
	"math"
	"sync"
)

type cacheblock struct {
	generation int
	data       []byte
}

type cache struct {
	mu         sync.RWMutex
	generation int
	maxblocks  int
	blocks     map[int64]cacheblock
}

func newcache() *cache {
	return &cache{
		generation: 0,
		maxblocks:  100,
		blocks:     make(map[int64]cacheblock),
	}
}

func (c *cache) evictold() {
	if len(c.blocks) < c.maxblocks {
		return
	}

	oldblk := int64(0)
	oldest := math.MaxInt32
	for off, blk := range c.blocks {
		if blk.generation < oldest {
			oldblk = off
			oldest = blk.generation
		}
	}

	delete(c.blocks, oldblk)
}

func (c *cache) put(offset int64, block []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.blocks[offset]
	if ok {
		return
	}

	gen := c.generation
	c.generation++
	c.blocks[offset] = cacheblock{gen, block}

	c.evictold()
}

func (c *cache) get(offset int64) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blk, ok := c.blocks[offset]
	if !ok {
		return nil
	}

	return blk.data
}
