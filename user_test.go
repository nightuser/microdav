package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type userData struct {
	Username string
	Password string
}

// TODO: separate into two functions: createTempFile & createDB because of defers
func createTempDB() (*os.File, *sql.DB, error) {
	var db *sql.DB = nil

	file, err := ioutil.TempFile(os.TempDir(), "test_users_*.db")
	if err != nil {
		return file, db, err
	}

	schema, err := ioutil.ReadFile("schema.sql")
	if err != nil {
		return file, db, err
	}

	db, err = sql.Open("sqlite3", file.Name())
	if err != nil {
		return file, db, err
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		return file, db, err
	}

	return file, db, nil
}

func TestCreateUser(t *testing.T) {
	file, err := createTempDB()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if err := createUser("testuser", "foo"); err != nil {
		t.Error(err)
	}
}

func TestDeleteUser(t *testing.T) {
	file, err := createTempDB()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if err := createUser("testuser", "foo"); err != nil {
		t.Error(err)
	}
	if err := deleteUser("testuser"); err != nil {
		t.Error(err)
	}
}

func TestDoubleCreateError(t *testing.T) {
	file, err := createTempDB()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	if err := createUser("testuser", "foo"); err != nil {
		t.Error(err)
	}
	err = createUser("testuser", "bar")
	if err == nil {
		t.Errorf("Second user creation must fail")
	}
	if err != ErrUserExists {
		t.Error(err)
	}
}

func TestGetUsers(t *testing.T) {
	file, err := createTempDB()
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	user1 := userData{
		Username: "user1",
		Password: "foo",
	}
	user2 := userData{
		Username: "user2",
		Password: "bar",
	}

	err = createUser(user1.Username, user1.Password)
	if err != nil {
		t.Error(err)
	}
	err = createUser(user2.Username, user2.Password)
	if err != nil {
		t.Error(err)
	}

	var user user
	rows, err := db.Query("select * from users")
	if err != nil {
		t.Error(err)
	}
	for rows.Next() {
		err := rows.Scan(&user.Username, &user.Salt, &user.Password)
		if err != nil {
			t.Error(err)
		}
		// TODO: add users to set and check
		fmt.Printf("%v\n", user)
	}
	rows.Close()
}

func testUsers() {
	if err := deleteUser("testuser"); err != nil {
		errorLogger.Fatal(err)
	}
	if err := createUser("testuser", "foo"); err != nil {
		errorLogger.Fatal(err)
	}
	if err := createUser("testuser", "bar"); err != nil {
		logger.Println("Adding user the second time failed; expected")
	}
}
