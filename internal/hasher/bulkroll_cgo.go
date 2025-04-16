//go:build cgo
// +build cgo

package hasher

/*
#include <stdint.h>

void buz_bulk_roll(uint8_t* buf, int len, int start, int window, int stride, uint64_t hash, const uint64_t* table, uint64_t* out) {
	int pos = start;
	int count = 0;

	while (pos + window <= len) {
		out[count++] = hash;

		for (int i = 0; i < stride; i++) {
			if (pos + window >= len) {
				return;
			}
			uint64_t outByte = table[buf[pos]];
			uint64_t inByte = table[buf[pos + window]];
			hash = (hash << 1 | hash >> 63) ^ (outByte << window | outByte >> (64 - window)) ^ inByte;
			pos++;
		}
	}
}
*/
import "C"
import (
	"unsafe"
)

func bulkRoll(buf []byte, start, windowSize, stride uint32, initialHash uint64) []uint64 {
	n := uint32(len(buf))
	if stride == 0 || windowSize == 0 || start+windowSize > n {
		return nil
	}

	capacity := (n-windowSize-start)/stride + 1
	hashes := make([]uint64, capacity)

	C.buz_bulk_roll(
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.int(len(buf)),
		C.int(start),
		C.int(windowSize),
		C.int(stride),
		C.uint64_t(initialHash),
		(*C.uint64_t)(unsafe.Pointer(&table)),
		(*C.uint64_t)(unsafe.Pointer(&hashes[0])),
	)

	return hashes
}
