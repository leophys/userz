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

func ParseOrder(ordBy, ordDir string) (Order, error) {
	var ord Order
	switch OrdBy(ordBy) {
	case OrdByFirstName:
		ord.OrdBy = OrdByFirstName
	case OrdByLastName:
		ord.OrdBy = OrdByLastName
	case OrdByNickName:
		ord.OrdBy = OrdByNickName
	case OrdByEmail:
		ord.OrdBy = OrdByEmail
	case OrdByUpdatedAt:
		ord.OrdBy = OrdByUpdatedAt
	case OrdByCreatedAt:
		ord.OrdBy = OrdByCreatedAt
	default:
		if ordBy == "" {
			ord.OrdBy = OrdByCreatedAt
		} else {
			return ord, fmt.Errorf("order_by not understood: %s", ordBy)
		}
	}

	switch ordDir {
	case "ASC", "":
		ord.OrdDir = OrdDirAsc
	case "DESC":
		ord.OrdDir = OrdDirDesc
	default:
		return ord, fmt.Errorf("order_dir not understood: %s", ordDir)
	}

	return ord, nil
}
