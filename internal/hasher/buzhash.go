package hasher

import (
	"encoding/binary"
	"errors"
	"hash"
	"math/bits"
)

const (
	hashSizeBytes = 8 // 64 bits = 8 bytes
)

var ErrNotWritable = errors.New("this hasher is not writable")
var ErrWindowTooLong = errors.New("the provided window size is longer that the buffer size")
var ErrIllegalRoll = errors.New("cannot roll any more")
var ErrIllegalStride = errors.New("illegal stride")

// 256 random uint64 numbers to map each byte.
var table = [256]uint64{
	0x6ac7ec3ab38484f, 0x925ae6f064171c28, 0x244495b205ce4c02, 0x66c6c503cd5c8230, 0x87882a1c9f1bc3e4, 0x979e21b3bf012cd2,
	0x6ff31e3b9a8b7056, 0x769db390094f640c, 0xe0a4a529b57bc3b, 0xc65560e187430aad, 0x52798fcb2fd23177, 0x64a792c71a9564d7,
	0xf39964707459aad1, 0xc879bee54d646279, 0x495d5c2cf4050ac3, 0x7d131a0421572566, 0xcab1d313d9b9365f, 0xbb566185972fe27f,
	0x3716a4e257a72420, 0x76e09e28e3ba151f, 0x7ae0cd88bf4022b8, 0x7eb8191af9ee2efa, 0xa69056756e4b209b, 0xa6d429436d97d1e5,
	0xa693b298e6009b6d, 0xb1fcfbbdb9c4968a, 0x23b026758294472d, 0x3db0dd6de3f13790, 0xf718b3bd1360ea86, 0xb76fa906c1bb53ce,
	0xa8aaa7ecff4d3c80, 0x78af34f2c2b211c8, 0x82af9cac5fdaf238, 0x539e69b70f0b9ce3, 0x8fb8c622f3de6b11, 0x1eb08a7aa9a97b5b,
	0x5210dabbb4b81de2, 0xff948b69a8803ba3, 0xd32c970b0dc60fd6, 0x6e364853438caafa, 0x46eee6f5d6d58e2e, 0xfbc987102182b39e,
	0x22cfb788dc34a095, 0x500530e71415124, 0xca9a1d57e5a64ea4, 0x187183fe5c5fae1e, 0xa4c04f966960636b, 0x6789f7bf3433b25a,
	0xe68b134541af4bbb, 0x2b12bdfda0cd2888, 0x57973822e755f2a0, 0xe0fd835d70218c66, 0x4f4e92692f24b98, 0x9aab82964927f460,
	0xe19d25c7b84d259b, 0x1981ae513780a89d, 0x225aa7aa7edc164d, 0x49b2bf05cf572e7d, 0x12c0dfe80eb7c01f, 0xb880fa65807ac921,
	0x98f683c79da6de7d, 0xe3e3f9712cb823e0, 0x112718fe7d6e0577, 0x965594a9083aeb59, 0xd23538e3f9a3a50d, 0xd281b4667af136de,
	0x6965a4268d1e0019, 0x5d5c5664a853523f, 0x51186d2df1ff305d, 0x5ca3afdddc7cd7d0, 0x7a1d198925e0df3, 0x252fd8d8eb2a48fe,
	0xc9783ed5d3178c6c, 0x4106755511fa7f64, 0x7c21cb8c25b489bc, 0x489459629f1c70b6, 0x2d9e41451a305457, 0xe9dd5eadac5b114b,
	0x8e16cb3284d3b3a1, 0x46de1e27b369a389, 0xe18c10897ac7a644, 0x1d74f7aeed69dbe1, 0x85086cfa562c3d51, 0x5249b1b0add0785a,
	0x62dee24b58981336, 0xdf9b86c236dedcf4, 0xa5a9699dd9911772, 0xcfcd3174dee306f7, 0x36f2a7a360f9b5d1, 0xde24e112854f5a86,
	0xc339740d4616b792, 0x23242cf1387c4434, 0xf863beef8a52a054, 0x6f57ac02b0805f82, 0x4ee02f75dd9d8b2b, 0x9db665e4ab9c8162,
	0x59463a08ff07c79c, 0xb054905d9a5189a9, 0xea105bcc1ced75a1, 0x9d3313bd02c4c785, 0xa1d1fe2c4e186f6a, 0xa451929c08d74f6f,
	0xc252dfb0a025b28a, 0x212aa981423ac4b5, 0x1accb6134653f0ea, 0x116a829730bef8fd, 0x134b78ca68261550, 0x3c81dc692faa174e,
	0x787c597ab1057102, 0xf262666a59f5bb77, 0x488c76b8d7dc71e6, 0x5b0927d6146c3a09, 0xe5e8392425368a87, 0xffa02d5f0db0020c,
	0x595c0f3384c97670, 0x4446900f000a21f8, 0xca5d93eff0b8ee1f, 0xa9ce4762daec734a, 0xf228a1a91eaff56b, 0x9a5dda6c513df5b,
	0xeeea75a99a70c05d, 0x2450181e7c778c1e, 0x3e20446fc0739a37, 0x789bd178a12432b8, 0x2cd013f2c6e30a02, 0x3e289c877e3216fb,
	0x2b467b97bdbbf046, 0x70a2e70c398385f9, 0xb309f50bedc45cd2, 0xab1c87318ed184cf, 0xff05174b6f671831, 0xaed32194063a4322,
	0x75429d35ddb5bfbd, 0x7ee1d1b65b71356e, 0x9bc6998b259da74d, 0x94d77784160c05e7, 0xe626bb219a8c0ff5, 0xd34e0ccc4ad855ac,
	0x91372ceceb7b8a24, 0xb86da3630e9c82eb, 0x4e45e3efcb170abb, 0x52e25ee2e2fa58d7, 0x3a29d311bca383d5, 0x849376b654ab150f,
	0xe88ec4da58b0d34, 0x7caa1b69024dd2a4, 0x70ad07192ba6926f, 0x1cf3d80c9c7fb676, 0xcefec8418d734bed, 0xb0b59760914b582b,
	0xfedf873bb570b5c9, 0x1e727de4bd865abc, 0xd1a8ceb82b920504, 0x88e8473859dd126f, 0xd212b3940fa5e6, 0xce029cce021d3ab8,
	0x472c692d26300943, 0xc3af132b79588a9c, 0x4185a9cc5ac2f073, 0x158f98b999ed69e5, 0x81cbb4e387cf5283, 0x5ff8a3a88663ff75,
	0xe2ca04ca27a2a31b, 0xb02c7c22030c25c8, 0x8d182d8a677ba196, 0xbdc1b7b1fd4b17aa, 0x31f294fc7a178d17, 0x45cab21ca66f5846,
	0xab04020244c4c4eb, 0xb9dcca0cf419981f, 0x2d91bed64347326a, 0x7218e987b5b1c051, 0x1165716693744a56, 0xd3d409c18c259aee,
	0xd5d676f27c86281f, 0xdd87ad06b53b2e87, 0x79e351c9dfb81de0, 0x5456880b83f75123, 0x55119453891f9dfa, 0xd8a362764f2c3732,
	0xbf1ad01073dee5ae, 0x9c223cab423829a5, 0x83c5fc36bd67266a, 0x390c9de142ec1ef0, 0xd833092ec8802130, 0x356dce302b7b1adb,
	0x3b4317cb6cfc66ec, 0x7ab780d895a87847, 0x5c9d6143c9d0ded2, 0x80ecee69d0ff12e3, 0xc373816880c77cdb, 0x54b7f418bb1e262,
	0xcf309f60ba38e5dc, 0x3bc2508d23438e21, 0xe1975e1332552dd, 0x58eec373f5101095, 0xb31830b2a0513c7b, 0x22cddb7b6a9ec6ef,
	0xf2becee9191551bb, 0xd88b42ebd98e8fd4, 0x92111f5561e7a8c6, 0xca7f3b6c4aad07ee, 0x3a8b89659cfbe7fc, 0x352402d921bbac38,
	0xb436ee0a191f9ddf, 0x43e9e16c8387a57, 0x851aab38ac27a0cb, 0x52168667f16bfd3d, 0x33f5bf131dfe8684, 0xe59d5b1bd51b025b,
	0xad9309e7234936b, 0x80bebf67a0e0615b, 0x82cda0f9170fa7f7, 0xebedaccbf0aef25e, 0xcd30a524e7a52b2, 0x3943adbdc665da65,
	0x8c535dcdd9ecd404, 0x3bb21c39e4c0d31b, 0x219ed8c92878321e, 0x4601c510eed8c6b7, 0xdfb017785619791e, 0xc16da260ea26c2e6,
	0xfe6ce35daf7c5ea2, 0x89d7aab27be3b408, 0x5790291fc841a55f, 0x1fa58028db6402cf, 0x74ca16d8de6a89ba, 0xcb87f3adc0e8ca03,
	0x7789849db69b71cf, 0x90d2b3ee093e7167, 0x42ea38f41ca84365, 0x5cf7d98dcb74536b, 0x79ac67f7a493ed7, 0x99876aa772dd632a,
	0xb3b4808dcfb682b7, 0xf0ba2ed5611261ad, 0x20560be7663813ed, 0x199b810be621dff9, 0x63667730b1f94cea, 0x7000eef34d6cb9a0,
	0x14696eba050883b5, 0x58015ab5b232d681, 0x783ba52bd84406f6, 0x3cf6688185746dce, 0xa111dedf4b43c846, 0xf74a7ada4264e8d,
	0xcec1d0c53f5e9c27, 0x2c62628525f7b6e1, 0x7c3da07729159f12, 0x86c3b06b44d6b976, 0x91c5c39aad05664a, 0xc04706984f5330b,
	0x731aba530128958c, 0x15ef4b57ffb49d88, 0xe89f7e30e4abc977, 0x3e8493ccd3ba8734,
}

// A quick helper to hash a fixed buffer such that `windowSize == len(buffer)`.
func Hash(buf []byte) uint64 {
	return hashBuf(buf)
}

// A rolling hash interface providing methods to be able to roll a hash
// forward in addition to the typical hash.Hash64 interface methods.
type RollingHash interface {
	hash.Hash64
	// Rolls the hasing window by the given step. Changes the window start position.
	Roll(step uint32) (uint64, error)
	// Rolls over the window at the given stride and returns all hashes.
	// Does not change the window starting position.
	BulkRoll(stride uint32) ([]uint64, error)
	// Get the current position in the input
	Position() uint32
}

// Implements RollingHash to calculate hashes rolling over a fixed buffer
// in steps or in bulk for better performance.
// Also implements hashing.Hash64 interface for interop but BEWARE that
// this is NOT a streaming hash and does not implement the incremental
// Write([]byte) method. The buffer has to be presented when constructing
// the object and can never be mutated. Reset() method will simply zero
// out all the state and make this object useless.
type Hasher struct {
	// The inner immutable buffer to hash over
	buf []byte
	// The window size for calculating the hash
	windowSize uint32
	// The current window start position
	position uint32
	// The current pre-computed hash
	hash uint64
}

// BulkRoll implements RollingHash.
func (h *Hasher) BulkRoll(stride uint32) ([]uint64, error) {
	if stride == 0 {
		return nil, ErrIllegalStride
	}

	return bulkRoll(h.buf, h.position, h.windowSize, stride, h.hash), nil
}

// Creates a new rolling hasher over the given buffer and window size the
// window starting from 0 index.
func New(buf []byte, windowSize uint32) (RollingHash, error) {
	if windowSize > uint32(len(buf)) {
		return nil, ErrWindowTooLong
	}

	return &Hasher{
		buf:        buf,
		windowSize: windowSize,
		position:   0,
		hash:       hashBuf(buf[:windowSize]),
	}, nil
}

// Inner method to hash the given bytes in one shot without rolling.
func hashBuf(p []byte) uint64 {
	var h uint64
	n := len(p)

	for i := 0; i < n; i++ {
		rot := n - 1 - i
		h ^= bits.RotateLeft64(table[p[i]], rot)
	}

	return h
}

// Rolls the hasing window by the given step. Changes the window start position.
func (h *Hasher) Roll(step uint32) (uint64, error) {
	// A full window must be present to be able to hash
	if h.position+uint32(step)+h.windowSize > uint32(len(h.buf)) {
		return 0, ErrIllegalRoll
	}

	for i := uint32(0); i < uint32(step); i++ {
		out := h.buf[h.position]
		in := h.buf[h.position+h.windowSize]

		h.hash = bits.RotateLeft64(h.hash, 1) ^
			bits.RotateLeft64(table[out], int(h.windowSize)) ^
			table[in]

		h.position++
	}

	return h.hash, nil
}

// Get the hash value of the current state of the hasher. Does not change the
// state in any way.
func (h *Hasher) Sum64() uint64 {
	return h.hash
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (h *Hasher) Sum(b []byte) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], h.Sum64())
	return append(b, buf[:]...)
}

// Reset the position of this hasher.
func (h *Hasher) Reset() {
	h.position = 0
	h.hash = hashBuf(h.buf[:h.windowSize])
}

// Size returns the number of bytes Sum will return.
func (h *Hasher) Size() int {
	return hashSizeBytes
}

// Not implemented and not applicable for this hash. The bytes are passed
// only with New and never updated.
func (h *Hasher) Write(p []byte) (int, error) {
	return 0, ErrNotWritable
}

// In buzhash context, a block size doesn't have any impact
func (h *Hasher) BlockSize() int {
	return 1
}

// Get the current position in the input
func (h *Hasher) Position() uint32 {
	return h.position
}
