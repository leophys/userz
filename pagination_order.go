package userz

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
