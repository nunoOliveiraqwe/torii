package store

import (
	"context"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/domain"
)

type UserStore interface {
	GetUserById(ctx context.Context, id int) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	GetRolesForUser(ctx context.Context, username string) ([]domain.Role, error)
	UpdateUser(user *domain.User, ctx context.Context) error
	InsertUser(user *domain.User, ctx context.Context) error
}
