package db

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dsn)
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(10)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	println("Connected to database")
	return db, nil
}
