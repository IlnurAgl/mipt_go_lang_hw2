package domain

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID       string
	Login    string
	Password string
}

func (u *User) HashPassword() error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

type UserRepository interface {
	GetUserByLogin(login string) (*User, error)
	CreateUser(user *User) error
}
