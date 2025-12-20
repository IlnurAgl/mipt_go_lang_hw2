package pg

import (
	"auth/internal/domain"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type UserPgRepository struct {
	db *sql.DB
}

func NewUserPgRepository(db *sql.DB) domain.UserRepository {
	return &UserPgRepository{db: db}
}

func (r *UserPgRepository) GetUserByLogin(login string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		"SELECT id, login, password FROM users WHERE login = $1",
		login,
	).Scan(&user.ID, &user.Login, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserPgRepository) CreateUser(user *domain.User) error {
	err := r.db.QueryRow(
		"INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id",
		user.Login, user.Password,
	).Scan(&user.ID)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.New("user already exists")
		}
		return err
	}

	return nil
}
