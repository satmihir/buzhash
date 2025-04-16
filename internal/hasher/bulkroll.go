//go:build !cgo
// +build !cgo

package hasher

import "math/bits"

func bulkRoll(buf []byte, start, windowSize, stride uint32, initialHash uint64) []uint64 {
	n := uint32(len(buf))
	capacity := (n-windowSize-start)/stride + 1
	hashes := make([]uint64, 0, capacity)

	pos := start
	hash := initialHash

	for {
		if pos+windowSize > n {
			break
		}

		hashes = append(hashes, hash)

		for i := uint32(0); i < stride; i++ {
			if pos+windowSize >= n {
				return hashes
			}
			out := buf[pos]
			in := buf[pos+windowSize]

			hash = bits.RotateLeft64(hash, 1) ^
				bits.RotateLeft64(table[out], int(windowSize)) ^
				table[in]

			pos++
		}
	}

	return hashes
}
