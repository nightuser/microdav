package usermanager

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const saltCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const saltLength = 8

var (
	// ErrUserExists is returned when user already exists during the
	// creation
	ErrUserExists = errors.New("User already exists")
	// ErrUserNotFound is returned when user not found in select/delete
	// query
	ErrUserNotFound = errors.New("User not found")
	// ErrPasswordDoesNotMatch is returned when the provided password is
	// wrong
	ErrPasswordDoesNotMatch = errors.New("Password doesn't match")
)

// UserCredentials holds username and password
type UserCredentials struct {
	Username string
	Password string
}

// User represents a database user entry
type User struct {
	UserCredentials
	Salt string
}

func saltPassword(password, salt string) []byte {
	return []byte(fmt.Sprintf("%s:%s", password, salt))
}

// UserManager provides encapsulated access to the users database
type UserManager struct {
	saltRand *rand.Rand
	db       *sql.DB
}

// New creates a UserManager
//
// Also see NewWithSeed
func New(dbName string) *UserManager {
	return NewWithSeed(dbName, time.Now().UnixNano())
}

// NewWithSeed behaves like New, but with the provided seed for the salt
// generator
func NewWithSeed(dbName string, seed int64) (manager *UserManager) {
	saltSource := rand.NewSource(seed)
	manager = &UserManager{
		saltRand: rand.New(saltSource),
	}
	return
}

// Close closes database connection
func (manager *UserManager) Close() error {
	if err := manager.db.Close(); err != nil {
		return err
	}

	return nil
}

func (manager *UserManager) generateSalt(n int) string {
	salt := make([]byte, n)
	length := len(saltCharset)
	for i := range salt {
		salt[i] = saltCharset[manager.saltRand.Intn(length)]
	}
	return string(salt)
}

// CheckPassword checks whether the provided credentials are correct
func (manager *UserManager) CheckPassword(username, password string) error {
	row := manager.db.QueryRow(
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

func (manager *UserManager) createUser(username string, password string) error {
	salt := manager.generateSalt(saltLength)
	saltedPassword := saltPassword(password, salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(
		saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = manager.db.Exec(
		`insert into users("username","salt","password") `+
			`values ($1, $2, $3)`,
		username, salt, hashedPassword)
	if err == sqlite3.ErrConstraintUnique {
		return ErrUserExists
	}
	return err
}

func (manager *UserManager) deleteUser(username string) error {
	result, err := manager.db.Exec(
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
