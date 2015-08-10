package httpreader

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const (
	testBlocks    = 128
	testBlockSize = 8192
)

var (
	testURL = "http://127.0.0.1:6789"
)

func init() {
	buf := make([]byte, testBlocks*testBlockSize)
	for i := 0; i < testBlocks; i++ {
		b := buf[i*testBlockSize:]
		copy(b, []byte(fmt.Sprintf("block%d", i)))
	}

	testBytes := bytes.NewReader(buf)

	handle := func(rw http.ResponseWriter, r *http.Request) {
		http.ServeContent(rw, r, "test", time.Now(), testBytes)
	}

	http.HandleFunc("/", handle)
	go http.ListenAndServe(":6789", nil)
}

func TestReaderAt(t *testing.T) {
	ra, err := NewReader(testURL)
	if err != nil {
		t.Fatalf("NewReader error: %v", err)
	}
	defer ra.Close()
	buf := make([]byte, testBlockSize)

	for i := 0; i < testBlocks; i++ {
		n, err := ra.ReadAt(buf, int64(i*testBlockSize))
		if err != nil {
			t.Errorf("ReadAt error: %v", err)
			return
		}

		if n != testBlockSize {
			t.Errorf("expected %d bytes got %d bytes", 512, n)
			return
		}

		exp := []byte(fmt.Sprintf("block%d", i))
		if !bytes.Contains(buf, exp) {
			t.Errorf("got %s want %s", buf, exp)
			return
		}
	}
}

func TestReader(t *testing.T) {
	ra, err := NewReader(testURL)
	if err != nil {
		t.Fatalf("NewReader error: %v", err)
	}
	defer ra.Close()

	// twice to hit cache
	for n := 0; n < 2; n++ {
		ra.Seek(0, 0)
		buf := make([]byte, testBlockSize)
		count := 0
		for i := 0; i < testBlocks; i++ {
			n, err := ra.Read(buf)
			if err != nil {
				t.Errorf("Read error: %v", err)
				return
			}
			if n != len(buf) {
				t.Errorf("expected %d bytes got %d bytes", 512, n)
				return
			}
			count += n
		}

		want := testBlocks * len(buf)

		if count != want {
			t.Errorf("short read, got %d want %d", count, want)
			return
		}
	}

	// check size
	sz, err := ra.Size()
	if err != nil {
		t.Errorf("stat error: %v", err)
		return
	}

	if sz != testBlocks*testBlockSize {
		t.Errorf("size got %d want %d", sz, testBlocks*testBlockSize)
		return
	}
}

func BenchmarkRead(b *testing.B) {
	ra, err := NewReader(testURL)
	if err != nil {
		b.Fatalf("NewReader error: %v", err)
	}
	buf := make([]byte, testBlockSize)

	for i := 0; i < b.N; i++ {
		n, err := ra.Read(buf)
		if err != nil {
			b.Fatal(err)
		}
		if n != len(buf) {
			b.Fatal("short read")
		}
	}

	b.SetBytes(int64(len(buf)))
}
