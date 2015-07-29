package httpreader

import (
	"bytes"
	"testing"
)

var (
	testURL string
)

func init() {
	testURL = "http://r-36.net/9front/9front.iso"
}

func TestReaderAt(t *testing.T) {
	ra := NewReader(testURL)
	buf := make([]byte, 512)

	n, err := ra.ReadAt(buf, 0x8000)
	if err != nil {
		t.Errorf("ReadAt error: %v", err)
		return
	}

	if n != 512 {
		t.Errorf("expected %d bytes got %d bytes", 512, n)
		return
	}

	exp := []byte("PLAN 9 FRONT")

	if !bytes.Contains(buf, exp) {
		t.Errorf("got %s want %s", buf, exp)
		return
	}
}

func TestReader(t *testing.T) {
	ra := NewReader(testURL)
	buf := make([]byte, 512)
	for i := 0; i < 8; i++ {
		n, err := ra.Read(buf)
		if err != nil {
			t.Errorf("Read error: %v", err)
			return
		}
		if n != 512 {
			t.Errorf("expected %d bytes got %d bytes", 512, n)
			return
		}
	}
}

func BenchmarkRead(b *testing.B) {
	ra := NewReader(testURL)
	buf := make([]byte, 8192)

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
