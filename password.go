package userz

import (
	"golang.org/x/crypto/bcrypt"
)

// Password represents a secret to be stored safely at rest.
type Password string

func NewPassword(plaintext string) (*Password, error) {
	return newPassword([]byte(plaintext))
}

func newPassword(plaintext []byte) (zero *Password, err error) {
	bres, err := bcrypt.GenerateFromPassword(plaintext, bcrypt.DefaultCost)
	if err != nil {
		return zero, err
	}

	res := Password(string(bres))
	return &res, nil
}

func (p *Password) Marshal(src []byte) error {
	res, err := newPassword(src)
	if err != nil {
		return err
	}

	p = res

	return nil
}

func (p *Password) Unmarshal() ([]byte, error) {
	return []byte(*p), nil
}

func (p *Password) String() string {
	return string(*p)
}
