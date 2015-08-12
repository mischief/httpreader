package cache

import (
	"container/list"
	"io"
	"sync"
)

type Cache struct {
	mu        sync.Mutex
	maxBlocks int
	blocks    map[int64]*cacheBlock
	lru       *list.List
	blockSize int
	size      int64
	reader    io.ReaderAt
}

type cacheBlock struct {
	lru    *list.Element
	offset int64
	data   []byte
}

// NewCache returns a new cache instance.
// Reads may be performed concurrently on the same cache,
// provided the reader supplied here is safe for concurrent usage.
func NewCache(blockSize, maxBlocks int, size int64, reader io.ReaderAt) *Cache {
	return &Cache{
		maxBlocks: maxBlocks,
		blocks:    make(map[int64]*cacheBlock),
		lru:       list.New(),
		blockSize: blockSize,
		size:      size,
		reader:    reader,
	}
}

// ReadAt reads the requested data via the cache,
// populating the cache with c.reader.ReadAt() as needed.
func (c *Cache) ReadAt(p []byte, offset int64) (n int, err error) {
	for ao := c.blockAlign(offset); n < len(p); ao += int64(c.blockSize) {
		blk := c.getBlock(ao, false)
		if blk == nil {
			// Clip the block size if necessary to c.size (EOF).
			// This prevents caching of growing files,
			// but simplifies EOF handling.
			bsize := c.blockSize
			if ao+int64(c.blockSize) > c.size {
				bsize = int(c.size - ao)
			}

			blk = &cacheBlock{offset: ao, data: make([]byte, bsize)}
			if x, err := c.reader.ReadAt(blk.data, ao); err != nil {
				return n + x, err
			}

			blk = c.addBlock(blk)
		}

		// Populate p with this chunk.
		n += copy(p[n:], blk.data[(offset+int64(n))-ao:])
	}
	return
}

// evictOld evicts the oldest block if c.maxBlocks is exceeded.
// c.mu must be locked.
func (c *Cache) evictOld() {
	if len(c.blocks) < c.maxBlocks {
		return
	}

	oldblk := c.lru.Remove(c.lru.Back()).(*cacheBlock)
	delete(c.blocks, oldblk.offset)
}

// addBlock adds the supplied cacheBlock to the cache,
// evicting an old one if needed.
// If the block is already cached, we simply return the cached instance.
func (c *Cache) addBlock(blk *cacheBlock) *cacheBlock {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cblk := c.getBlock(blk.offset, true); cblk != nil {
		return cblk
	}

	c.blocks[blk.offset] = blk
	blk.lru = c.lru.PushFront(blk)
	c.evictOld()

	return blk
}

// getBlock gets the block at offset from the cache.
// Every time a block is gotten via the cache, it's moved to the lru front.
func (c *Cache) getBlock(aoffset int64, locked bool) *cacheBlock {
	if !locked {
		c.mu.Lock()
		defer c.mu.Unlock()
	}

	if blk, ok := c.blocks[aoffset]; ok {
		c.lru.MoveToFront(blk.lru)
		return blk
	}

	return nil
}

// blockAlign aligns offset to the cache alignment (blockSize)
func (c *Cache) blockAlign(offset int64) int64 {
	return offset / int64(c.blockSize) * int64(c.blockSize)
}
