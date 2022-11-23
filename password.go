package userz

import (
    "encoding/json"
)

// Password represents a secret to be stored safely at rest.
// Its plaintext should only live in memory, and never be persisted.
type Password interface {
    Encrypt(salt, plaintext string) (secret []byte)
    Decrypt(secret []byte) (salt string, plaintext string)

    json.Marshaler
    json.Unmarshaler
}
