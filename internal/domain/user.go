package domain

import (
	"time"
)

type User struct {
	ID               int
	Username         string
	Password         string
	IsFirstTimeLogin bool
	Active           bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewUser(username, password string) *User {
	return &User{
		Username:         username,
		Password:         password,
		Active:           false,
		IsFirstTimeLogin: true,
	}
}

type UserFilter struct {
	Username *string
	ID       *int
	Limit    int
	Offset   int
}
