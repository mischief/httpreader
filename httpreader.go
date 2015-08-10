package httpreader

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Reader reads files served over http/https from servers supporting http range
// requests.
type Reader struct {
	client *http.Client
	url    string
	offset int64
	size   int64

	cache *Cache
}

func NewReader(url string) (*Reader, error) {
	ra := &Reader{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		url:    url,
		offset: 0,
	}

	size, err := ra.getSize()
	if err != nil {
		return nil, err
	}

	ra.size = size
	ra.cache = NewCache(32768, 100, size)

	return ra, nil
}

// ReadAt implements io.ReaderAt.
func (ra *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	return ra.cache.Get(p, off, &fetcher{ra: ra})
}

// Read implements io.Reader.
func (ra *Reader) Read(p []byte) (n int, err error) {
	n, err = ra.ReadAt(p, ra.offset)
	if err == nil {
		ra.offset += int64(n)
	}
	return
}

// Close implements io.Closer.
func (ra *Reader) Close() error {
	return nil
}

// Seek implements io.Seeker
func (ra *Reader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		ra.offset = offset
	case 1:
		ra.offset += offset
	case 2:
		sz, err := ra.Size()
		if err != nil {
			return 0, err
		}
		ra.offset = sz + offset
	default:
		return 0, fmt.Errorf("invalid whence %d", whence)
	}
	return ra.offset, nil
}

// Size returns the size of the content.
func (ra *Reader) Size() (int64, error) {
	return ra.size, nil
}

// getSize returns the size of the content. It will return an error if
// the remote server does not support HEAD requests or the
// Content-Length header.
func (ra *Reader) getSize() (int64, error) {
	// stat remote file, make sure it's seekable
	resp, err := ra.client.Head(ra.url)
	if err != nil {
		return 0, err
	}
	cls := resp.Header.Get("Content-Length")
	if cls == "" {
		return 0, fmt.Errorf("unseekable content")
	}

	var n int64
	fmt.Sscan(cls, &n)
	return n, nil
}

var (
	_ io.Reader   = &Reader{}
	_ io.Closer   = &Reader{}
	_ io.Seeker   = &Reader{}
	_ io.ReaderAt = &Reader{}
)

// fetcher fulfills a ReaderAt interface for the cache populator.
type fetcher struct {
	ra *Reader
}

func (f *fetcher) ReadAt(p []byte, off int64) (n int, err error) {
	req, err := http.NewRequest("GET", f.ra.url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, off+int64(len(p))))

	resp, err := f.ra.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return io.ReadFull(resp.Body, p)
}
