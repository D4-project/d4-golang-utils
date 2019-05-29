package hash

import (
	uuid "github.com/gofrs/uuid"
)

// NewV4 returns a randomly generated UUID.
func NewV4() (uuid.UUID, error) {
	return uuid.NewV4()
}

// FromString returns a UUID parsed from the input string.
func FromString(input string) (uuid.UUID, error) {
	return uuid.FromString(input)
}

// FromBytesOrNil returns a UUID generated from the raw byte slice input.
func FromBytesOrNil(input []byte) uuid.UUID {
	return uuid.FromBytesOrNil(input)
}
