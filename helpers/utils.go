package helpers

import (
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
