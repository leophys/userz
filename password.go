package userz

import (
	"golang.org/x/crypto/bcrypt"
)

// Password represents a secret to be stored safely at rest.
type Password []byte

func NewPassword(plaintext string) (Password, error) {
	return newPassword([]byte(plaintext))
}

func newPassword(plaintext []byte) (zero Password, err error) {
	hashed, err := bcrypt.GenerateFromPassword(plaintext, bcrypt.DefaultCost)
	if err != nil {
		return zero, err
	}

	return []byte(hashed), nil
}

func (p Password) String() string {
	return string(p)
}

type Passworder func(string) (Password, error)
