package random

import (
	"crypto/rand"
	"math"
	"math/big"
)

// NumCode generates random "length"-digit code
func NumCode(length uint) (uint64, error) {
	lengthInt := int(length)

	if length <= 1 {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return 0, err
		}

		return n.Uint64(), nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(9*int64(math.Pow10(lengthInt-1))))
	if err != nil {
		return 0, err
	}

	return n.Uint64() + uint64(math.Pow10(lengthInt-1)), nil
}
