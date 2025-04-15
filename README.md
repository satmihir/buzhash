# buzhash

A blazing-fast, resource frugal rolling hash library in pure Go.

Inspired by the original [BuzHash](https://en.wikipedia.org/wiki/Rolling_hash#BuzHash) design, this implementation:

- Is optimized for fixed-length phrase hashing and sliding window roll-forward
- Supports both one-shot and rolling window APIs
- Implements hash.Hash64 for interoperability
- Designed for use in tools like string detectors, phrase indexers, and content filters

## Install

```
go get github.com/satmihir/buzhash
```

## Quick Example

```go
import "github.com/satmihir/buzhash"

// Hash a single buffer
hash := buzhash.Hash([]byte("foobar"))

// Create a rolling hasher
h, err := buzhash.New([]byte("hello world"), 4)
if err != nil {
    log.Fatal(err)
}

// Roll forward by 1 byte
next, _ := h.Roll(1)

// Get multiple window hashes with stride 1
hashes, _ := h.BulkRoll(1)
```

## Benchmark

Tested with window size 6 against "Pride and Prejudice" book from project Gutenberg - 754kb.

Tested on macOS (Apple M1 Pro, Go 1.22):

| Function      | Data Size | Time/op | MB/s     | Allocs/op | Mem/op |
| ------------- | --------- | ------- | -------- | --------- | ------ |
| `Roll(1)`     | 754 KB    | 2.84 ms | 265 MB/s | 0         | 0 B    |
| `BulkRoll(1)` | 754 KB    | 1.72 ms | 437 MB/s | 1         | 6.0 MB |

## License

MIT
