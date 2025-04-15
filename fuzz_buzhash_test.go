package buzhash

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func FuzzRollingCorrectness(f *testing.F) {
	// Seed corpus with a few fixed cases
	f.Add([]byte("hello world"), uint32(3))
	f.Add([]byte("abc"), uint32(2))
	f.Add([]byte("1234567890abcdef"), uint32(4))

	f.Fuzz(func(t *testing.T, data []byte, window uint32) {
		if len(data) == 0 || window == 0 || int(window) > len(data) {
			return // invalid setup
		}

		h, err := New(data, window)
		if err != nil {
			return // skip invalid combo
		}

		// Verify Roll(1) matches Hash(buf[i:i+window])
		var rollingHashes []uint64
		rollingHashes = append(rollingHashes, h.Sum64())

		for i := 1; i+int(window) <= len(data); i++ {
			hash, err := h.Roll(1)
			assert.NoError(t, err)
			rollingHashes = append(rollingHashes, hash)
		}

		// Ground truth
		for i := 0; i+int(window) <= len(data); i++ {
			expected := Hash(data[i : i+int(window)])
			assert.Equal(t, expected, rollingHashes[i], "mismatch at offset %d", i)
		}

		// Reset and rerun — must match again
		h.Reset()
		assert.Equal(t, Hash(data[:window]), h.Sum64())
	})
}
