package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose/v3"

	"movie-service/migrations"
)

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(migrations.Files)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set migration dialect: %w", err)
	}
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
