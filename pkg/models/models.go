package models

import (
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")
var ErrDuplicateEmail = errors.New("models: duplicate email")
var ErrInvalidCredentials = errors.New("models: invalid credentials")

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
	Active         bool
}

type Thread struct {
	ID      int
	Topic   string
	User    *User
	Created time.Time
	Posts   []*Post
}

type Post struct {
	ID      int
	Body    string
	User    *User
	Created time.Time
}
