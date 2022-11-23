package userz

import (
    "bytes"
    "io/ioutil"
    "encoding/json"
    "testing"

	"golang.org/x/crypto/bcrypt"
    "github.com/stretchr/testify/assert"
)

const testPass = "ciaomiaobau"

type testVessel struct {
    Password *Password `json:"password"`
}

func TestPassword(t *testing.T) {
    assert := assert.New(t)

    b := bytes.NewBuffer(nil)
    enc := json.NewEncoder(b)

    pass, err := NewPassword(testPass)
    assert.NoError(err)

    err = enc.Encode(&testVessel{
        Password: pass,
    })
    assert.NoError(err)

    w, err := ioutil.ReadAll(b)
    assert.NoError(err)

    var newVessel testVessel

    err = json.Unmarshal([]byte(w), &newVessel)
    assert.NoError(err)

    err = bcrypt.CompareHashAndPassword([]byte(*newVessel.Password), []byte(testPass))
    assert.NoError(err)
}
