package users

type IUsers interface {
	New() (IUser, error)
	Get(id string) IUser
}

type IUser interface {
	ID() string
	Data() map[string]interface{}
}

type user struct {
	id   string
	data map[string]interface{}
}
