package httpreader

import (
	"bytes"
	"testing"
)

func TestCache(t *testing.T) {
	c := newcache()

	for i := 0; i < 150; i++ {
		b := bytes.Repeat([]byte{byte(i)}, 512)
		c.put(int64(i*512), b)
	}

	// check for eviction
	buf := c.get(0)
	if buf != nil {
		t.Errorf("expected nil got %v", buf)
		return
	}
}