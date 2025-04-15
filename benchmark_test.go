package buzhash

import (
	"os"
	"testing"
)

var result uint64

func loadBook(b *testing.B) []byte {
	data, err := os.ReadFile("testdata/book.txt")
	if err != nil {
		b.Fatalf("failed to load test book: %v", err)
	}
	return data
}

func BenchmarkRollOneByOne(b *testing.B) {
	data := loadBook(b)
	window := uint32(6)

	h, err := New(data, window)
	if err != nil {
		b.Fatalf("failed to init hasher: %v", err)
	}

	var last uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		for j := 1; j+int(window) <= len(data); j++ {
			last, err = h.Roll(1)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
	result = last
}

func BenchmarkBulkRoll(b *testing.B) {
	data := loadBook(b)
	window := uint32(6)

	h, err := New(data, window)
	if err != nil {
		b.Fatalf("failed to init hasher: %v", err)
	}

	var hashes []uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		hashes, err = h.BulkRoll(1)
		if err != nil {
			b.Fatal(err)
		}
	}
	if len(hashes) > 0 {
		result = hashes[len(hashes)-1]
	}
}
