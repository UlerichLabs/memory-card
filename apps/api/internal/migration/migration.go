package migration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/UlerichLabs/memory-card/apps/api/migrations"
)

func Apply(pool *pgxpool.Pool) (bool, error) {
	database := stdlib.OpenDBFromPool(pool)
	defer database.Close()
	return ApplyDatabase(database)
}

func ApplyDatabase(database *sql.DB) (bool, error) {
	driver, err := pgxmigrate.WithInstance(database, &pgxmigrate.Config{})
	if err != nil {
		return false, fmt.Errorf("criar driver pgx para migrate: %w", err)
	}
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return false, fmt.Errorf("carregar migrations do embed.FS: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return false, fmt.Errorf("instanciar migrate: %w", err)
	}
	defer func() { _, _ = m.Close() }()
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return false, nil
		}
		return false, fmt.Errorf("executar migrate up: %w", err)
	}
	return true, nil
}
