package httpreader

import (
	"container/list"
	"sync"
)

type cacheblock struct {
	lru    *list.Element
	offset int64
	data   []byte
}

type cache struct {
	mu        sync.RWMutex
	maxblocks int
	blocks    map[int64]*cacheblock
	lru       *list.List
}

func newcache() *cache {
	return &cache{
		maxblocks: 100,
		blocks:    make(map[int64]*cacheblock),
		lru:       list.New(),
	}
}

func (c *cache) evictold() {
	if len(c.blocks) < c.maxblocks {
		return
	}

	oldblk := c.lru.Remove(c.lru.Back()).(*cacheblock)
	delete(c.blocks, oldblk.offset)
}

func (c *cache) put(offset int64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.blocks[offset]
	if ok {
		return
	}

	blk := cacheblock{offset: offset, data: data}
	c.blocks[offset] = &blk
	blk.lru = c.lru.PushFront(&blk)

	c.evictold()
}

func (c *cache) get(offset int64) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blk, ok := c.blocks[offset]
	if !ok {
		return nil
	}

	c.lru.MoveToFront(blk.lru)

	return blk.data
}
