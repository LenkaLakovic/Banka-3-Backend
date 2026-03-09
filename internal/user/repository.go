package user

type User struct {
	id             uint64
	hashedPassword string
}

func GetUserByEmail(id string) User {
	panic("not implemented yet!!!")
}
