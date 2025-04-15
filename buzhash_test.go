package buzhash

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash64Interface(t *testing.T) {
	data := []byte("hello world")
	windowSize := uint32(5)

	_, err := New(data, uint32(len(data)+1))
	assert.Error(t, err)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	// Confirm it satisfies hash.Hash64
	var _ hash.Hash64 = h

	// Sum64 must equal a full re-hash of first window
	expected := h.Sum64()
	expectedDirect := Hash(data[:windowSize])
	assert.Equal(t, expected, expectedDirect)

	// Sum appends bytes properly
	base := []byte{1, 2, 3}
	out := h.Sum(base)
	assert.True(t, bytes.Equal(out[:3], base))
	assert.Equal(t, len(out), 11)

	sumBytes := out[3:]
	got := binary.BigEndian.Uint64(sumBytes)
	assert.Equal(t, got, expected)

	// Write should always error
	n, err := h.Write([]byte("abc"))
	assert.True(t, errors.Is(err, ErrNotWritable))
	assert.Equal(t, n, 0)
	assert.Equal(t, h.Size(), 8)
	assert.Equal(t, h.BlockSize(), 1)
}

func TestRoll(t *testing.T) {
	data := []byte("abcdefghijk")
	windowSize := uint32(4)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	// Store expected initial hash
	expected := []uint64{
		h.Sum64(),
	}

	// Roll one at a time and collect rolling hashes
	for i := 1; i <= len(data)-int(windowSize); i++ {
		hash, err := h.Roll(1)
		assert.NoError(t, err)
		expected = append(expected, hash)
	}

	// Now recompute expected hashes manually
	for i := 0; i <= len(data)-int(windowSize); i++ {
		sub := data[i : i+int(windowSize)]
		want := Hash(sub)
		assert.Equal(t, want, expected[i], "rolling hash mismatch at offset %d", i)
	}

	// Try illegal roll (should exceed buffer)
	_, err = h.Roll(1)
	assert.ErrorIs(t, err, ErrIllegalRoll)
}

func TestBulkRollStride1(t *testing.T) {
	data := []byte("abcdefghijk")
	windowSize := uint32(3)
	stride := uint32(1)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	// Save original position
	origPos := h.Sum64()

	hashes, err := h.BulkRoll(stride)
	assert.NoError(t, err)

	// Compute expected manually
	var expected []uint64
	for i := 0; i+int(windowSize) <= len(data); i += int(stride) {
		expected = append(expected, Hash(data[i:i+int(windowSize)]))
	}

	assert.Equal(t, expected, hashes, "bulk roll hashes do not match expected values")

	// Ensure position has not changed
	assert.Equal(t, origPos, h.Sum64(), "BulkRoll should not mutate internal state")

	// Test illegal stride
	_, err = h.BulkRoll(0)
	assert.ErrorIs(t, err, ErrIllegalStride)
}

func TestBulkRollStride2(t *testing.T) {
	data := []byte("abcdefghijkx")
	windowSize := uint32(3)
	stride := uint32(2)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	// Save original position
	origPos := h.Sum64()

	hashes, err := h.BulkRoll(stride)
	assert.NoError(t, err)

	// Compute expected manually
	var expected []uint64
	for i := 0; i+int(windowSize) <= len(data); i += int(stride) {
		expected = append(expected, Hash(data[i:i+int(windowSize)]))
	}

	assert.Equal(t, expected, hashes, "bulk roll hashes do not match expected values")

	// Ensure position has not changed
	assert.Equal(t, origPos, h.Sum64(), "BulkRoll should not mutate internal state")

	// Test illegal stride
	_, err = h.BulkRoll(0)
	assert.ErrorIs(t, err, ErrIllegalStride)
}

func TestRoll_MultiStep(t *testing.T) {
	data := []byte("abcdefghijk")
	windowSize := uint32(3)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	// Roll forward by 2
	hash2, err := h.Roll(2)
	assert.NoError(t, err)
	assert.Equal(t, Hash(data[2:5]), hash2)

	// Position should now be 2
	assert.Equal(t, data[2:5], data[h.Position():int(h.Position())+int(windowSize)])

	// Roll forward again by 1 (total offset now 3)
	hash3, err := h.Roll(1)
	assert.NoError(t, err)
	assert.Equal(t, Hash(data[3:6]), hash3)

	// Try rolling past end — should error
	_, err = h.Roll(uint32(len(data)))
	assert.ErrorIs(t, err, ErrIllegalRoll)
}

func TestReset(t *testing.T) {
	data := []byte("abcdefghijk")
	windowSize := uint32(4)

	h, err := New(data, windowSize)
	assert.NoError(t, err)

	originalHash := h.Sum64()

	// Roll forward a couple times
	_, err = h.Roll(2)
	assert.NoError(t, err)

	// Hash should now be different
	assert.NotEqual(t, originalHash, h.Sum64(), "Hash should change after rolling")

	// Reset
	h.Reset()

	// Hash should match original again
	assert.Equal(t, originalHash, h.Sum64(), "Hash should return to original after Reset")

	// Roll again and ensure it behaves the same as the first time
	hash, err := h.Roll(1)
	assert.NoError(t, err)
	assert.Equal(t, Hash(data[1:1+int(windowSize)]), hash)

	// Call Reset again to make sure it doesn't panic or break
	h.Reset()
	assert.Equal(t, originalHash, h.Sum64(), "Reset should be idempotent")
}
