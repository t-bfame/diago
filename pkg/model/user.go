package model

type UserID string

type User struct {
	ID       UserID
	Username string
	Password string
}
