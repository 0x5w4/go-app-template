package shared

import (
	"errors"

	"github.com/google/uuid"
)

func GenerateUUID() (uuid.UUID, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	return newUUID, nil
}

func GenerateUUIDString() (string, error) {
	u, err := GenerateUUID()
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func ParseUUID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.New("invalid UUID format")
	}

	return u, nil
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func GenerateUUIDv5(namespace uuid.UUID, name string) uuid.UUID {
	return uuid.NewSHA1(namespace, []byte(name))
}
