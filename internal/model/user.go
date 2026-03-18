package model

import (
	"context"
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

type UserDal interface {
	GetUserById(ctx context.Context, id int) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetRolesForUser(ctx context.Context, username string) ([]Role, error)
	UpdateUser(user *User, ctx context.Context) error
	InsertUser(user *User, ctx context.Context) error
}

type UserFilter struct {
	Username *string
	ID       *int
	Limit    int
	Offset   int
}
