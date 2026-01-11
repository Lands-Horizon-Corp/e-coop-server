package helpers

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

func GenerateToken() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", eris.Wrap(err, "token generation failed")
	}
	return id.String(), nil
}

func GenerateDigitCode(digits int) (string, error) {
	if digits <= 0 {
		return "", eris.New("digits must be greater than 0")
	}
	n, err := rand.Int(rand.Reader, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil))
	if err != nil {
		return "", eris.Wrap(err, "digit code generation failed")
	}
	return fmt.Sprintf(fmt.Sprintf("%%0%dd", digits), n.Int64()), nil
}
