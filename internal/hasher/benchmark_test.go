package hasher

import (
	"fmt"
	"math"
	"math/bits"
	"os"
	"testing"

	"github.com/spaolacci/murmur3"
)

var result uint64

func loadBook() []byte {
	data, _ := os.ReadFile("testdata/book.txt")
	return data
}

func BenchmarkRollOneByOne(b *testing.B) {
	data := loadBook()
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
	data := loadBook()
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

func BenchmarkRecomputeEveryWindow(b *testing.B) {
	data := loadBook()
	window := 6
	n := len(data) - window + 1

	var h uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			h = Hash(data[j : j+window])
		}
	}
	result = h
}

func BenchmarkMurmur3EveryWindow(b *testing.B) {
	data := loadBook()
	window := 6
	n := len(data) - window + 1

	var h uint64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			h = murmur3.Sum64(data[j : j+window])
		}
	}
	result = h
}

func extractRollingHashes(buf []byte, window int, useBuzhash bool) []uint64 {
	var hashes []uint64
	n := len(buf) - window + 1
	if useBuzhash {
		for i := 0; i < n; i++ {
			h := Hash(buf[i : i+window])
			hashes = append(hashes, h)
		}
	} else {
		for i := 0; i < n; i++ {
			h := murmur3.Sum64(buf[i : i+window])
			hashes = append(hashes, h)
		}
	}
	return hashes
}

func analyzeDistribution(hashes []uint64, bins int) {
	var bucket = make([]int, bins)
	for _, h := range hashes {
		bucket[h%uint64(bins)]++
	}

	mean := float64(len(hashes)) / float64(bins)
	var variance float64
	for _, count := range bucket {
		diff := float64(count) - mean
		variance += diff * diff
	}
	variance /= float64(bins)
	fmt.Printf("\nUniformity (mod %d):\n", bins)
	//for i, count := range bucket {
	//	fmt.Printf("Bin %2d: %6d\n", i, count)
	//}
	fmt.Printf("Mean: %.2f, StdDev: %.2f\n", mean, math.Sqrt(variance))
}

func bitEntropy(hashes []uint64) {
	var ones [64]int
	for _, h := range hashes {
		for i := 0; i < 64; i++ {
			if (h>>i)&1 == 1 {
				ones[i]++
			}
		}
	}
	fmt.Println("\nBit entropy per bit:")
	for i := 0; i < 64; i++ {
		p0 := float64(len(hashes)-ones[i]) / float64(len(hashes))
		p1 := float64(ones[i]) / float64(len(hashes))
		h := 0.0
		if p0 > 0 {
			h -= p0 * math.Log2(p0)
		}
		if p1 > 0 {
			h -= p1 * math.Log2(p1)
		}
		fmt.Printf("Bit %2d: entropy = %.4f\n", i, h)
	}
}

func trailingZeroStats(hashes []uint64) {
	maxZeros := 0
	for _, h := range hashes {
		z := bits.TrailingZeros64(h)
		if z > maxZeros {
			maxZeros = z
		}
	}
	bins := make([]int, maxZeros+1)
	for _, h := range hashes {
		bins[bits.TrailingZeros64(h)]++
	}
	fmt.Println("\nTrailing zero distribution:")
	for i, count := range bins {
		fmt.Printf("%2d zeros: %6d\n", i, count)
	}
}

func TestEntropyComparison(t *testing.T) {
	// Using book test
	book := loadBook()
	window := 6

	types := []struct {
		name string
		hash bool
	}{
		{"BuzHash", true},
		{"Murmur3", false},
	}

	for _, typ := range types {
		fmt.Printf("\n===== %s =====\n", typ.name)
		hashes := extractRollingHashes(book, window, typ.hash)
		analyzeDistribution(hashes, 65536)
		bitEntropy(hashes)
		trailingZeroStats(hashes)
	}
}
