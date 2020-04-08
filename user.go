package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const saltCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const saltLength = 8

var saltRand *rand.Rand

var (
	// ErrUserExists is returned when user already exists during the
	// creation
	ErrUserExists = errors.New("User already exists")
	// ErrUserNotFound is returned when user not found in select/delete
	// query
	ErrUserNotFound = errors.New("User not found")
	// ErrPasswordDoesNotMatch is return when the provided password is wrong
	ErrPasswordDoesNotMatch = errors.New("Password doesn't match")
)

type user struct {
	Username string
	Salt     string
	Password string
}

func checkPassword(username, password string) error {
	row := db.QueryRow(
		`select "salt", "password" from users `+
			`where "username" = $1`,
		username)
	var salt string
	var hashedPassword []byte
	if err := row.Scan(&salt, &hashedPassword); err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	saltedPassword := saltPassword(password, salt)
	if err := bcrypt.CompareHashAndPassword(
		hashedPassword, saltedPassword); err != nil {
		return ErrPasswordDoesNotMatch
	}

	return nil
}

func generateSalt(n int) string {
	salt := make([]byte, n)
	length := len(saltCharset)
	for i := range salt {
		salt[i] = saltCharset[saltRand.Intn(length)]
	}
	return string(salt)
}

func saltPassword(password, salt string) []byte {
	return []byte(fmt.Sprintf("%s:%s", password, salt))
}

func createUser(username string, password string) error {
	salt := generateSalt(saltLength)
	saltedPassword := saltPassword(password, salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(
		saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		`insert into users("username","salt","password") `+
			`values ($1, $2, $3)`,
		username, salt, hashedPassword)
	if err == sqlite3.ErrConstraintUnique {
		return ErrUserExists
	}
	return err
}

func deleteUser(username string) error {
	result, err := db.Exec(
		`delete from users `+
			`where "username" = $1`,
		username)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return ErrUserNotFound
	}

	return nil
}
