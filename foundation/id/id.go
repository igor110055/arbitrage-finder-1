package id

import (
	"crypto/rand"
	"github.com/oklog/ulid/v2"
	"math"
	"sync"
)

var entropyPool = sync.Pool{
	New: func() interface{} {
		return ulid.Monotonic(rand.Reader, math.MaxUint32)
	},
}

func ULID() ulid.ULID {
	entropy, _ := entropyPool.Get().(*ulid.MonotonicEntropy)
	defer entropyPool.Put(entropy)

	return ulid.MustNew(ulid.Now(), entropy)
}
