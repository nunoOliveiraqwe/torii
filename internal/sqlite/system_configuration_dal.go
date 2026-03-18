package sqlite

import (
	"context"

	"github.com/nunoOliveiraqwe/micro-proxy/internal/model"
)

// Ensure service implements interface.
var _ model.SystemConfigurationDal = (*SystemConfigurationSqliteDal)(nil)

type SystemConfigurationSqliteDal struct {
	db *DB
}

func NewSystemConfigurationSqliteDal(db *DB) model.SystemConfigurationDal {
	return &SystemConfigurationSqliteDal{db: db}
}

func (s *SystemConfigurationSqliteDal) GetSystemConfiguration() (*model.SystemConfiguration, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var config model.SystemConfiguration
	err = tx.QueryRowContext(ctx, `
		SELECT
			id,
			first_time_setup_complete
		FROM system_configuration
		WHERE id = 1`,
	).Scan(
		&config.ID,
		&config.IsFirstTimeSetupConcluded,
	)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *SystemConfigurationSqliteDal) UpdateSystemConfiguration(config *model.SystemConfiguration) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := updateSystemConfiguration(ctx, tx, config); err != nil {
		return err
	}
	return tx.Commit()
}

func updateSystemConfiguration(ctx context.Context, tx *Tx, config *model.SystemConfiguration) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE system_configuration SET
			first_time_setup_complete = ?
		WHERE id = 1`,
		config.IsFirstTimeSetupConcluded,
	)
	return err
}
