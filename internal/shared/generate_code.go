package shared

import (
	"fmt"
	"math"
	rand "math/rand/v2"

	"github.com/cockroachdb/errors"
)

func GenerateCode(prefix string, length int) (string, error) {
	if len(prefix) != 2 {
		return "", errors.New("prefix must be exactly 2 characters")
	}

	if length <= 0 || length > 18 {
		return "", errors.New("length must be between 1 and 18")
	}

	upperBound := int64(math.Pow10(length))
	randomNumber := rand.Int64N(upperBound)

	format := fmt.Sprintf("%%0%dd", length)
	randomNumberStr := fmt.Sprintf(format, randomNumber)

	return prefix + randomNumberStr, nil
}
