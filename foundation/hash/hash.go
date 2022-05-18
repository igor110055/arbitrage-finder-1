package hash

import (
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
)

func GenerateHash(str string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(str)); err != nil {
		panic(errors.Wrap(err, "failed to generate string hash"))
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
