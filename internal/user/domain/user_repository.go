package domain

type UserRepository interface {
	Save(user *User) error
	FindById(id int64) (*User, error)
	Update(user *User) error
	DeleteById(id int64) error
}
