package model

import "context"

type Role struct {
	ID   int
	Name string
}

type RoleDal interface {
	GetRoleById(ctx context.Context, id int) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
}
