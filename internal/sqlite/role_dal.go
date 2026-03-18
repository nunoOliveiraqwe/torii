package sqlite

import (
	"context"
	"fmt"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/model"
)

// Ensure service implements interface.
var _ model.RoleDal = (*RoleSqliteDal)(nil)

type RoleSqliteDal struct {
	db *DB
}

func NewRoleSqliteDal(db *DB) model.RoleDal {
	return &RoleSqliteDal{db: db}
}

func (s *RoleSqliteDal) GetRoleById(ctx context.Context, id int) (*model.Role, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var role model.Role
	err = tx.QueryRowContext(ctx, `
		SELECT id, name
		FROM role
		WHERE id = ?`,
		id,
	).Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by id %d: %w", id, err)
	}

	return &role, nil
}

func (s *RoleSqliteDal) GetRoleByName(ctx context.Context, name string) (*model.Role, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var role model.Role
	err = tx.QueryRowContext(ctx, `
		SELECT id, name
		FROM role
		WHERE name = ?`,
		name,
	).Scan(&role.ID, &role.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get role by name %s: %w", name, err)
	}

	return &role, nil
}
