package userz

import (
	"fmt"
)

type OrdBy string

const (
	OrdByFirstName OrdBy = "first_name"
	OrdByLastName  OrdBy = "last_name"
	OrdByNickName  OrdBy = "nick_name"
	OrdByEmail     OrdBy = "email"
	OrdByCreatedAt OrdBy = "created_at"
	OrdByUpdatedAt OrdBy = "updated_at"
)

func (o OrdBy) String() string {
	return string(o)
}

type OrdDir bool

const (
	OrdDirDesc OrdDir = true
	OrdDirAsc  OrdDir = false
)

func (o OrdDir) String() string {
	if bool(o) {
		return "DESC"
	}
	return "ASC"
}

type Order struct {
	OrdBy
	OrdDir
}

func (o Order) String() string {
	return fmt.Sprintf("%s %s", o.OrdBy, o.OrdDir)
}
