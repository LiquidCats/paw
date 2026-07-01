package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rotisserie/eris"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(ctx context.Context, conn *pgx.Conn) error {
	defer func() {
		_ = conn.Close(ctx)
	}()

	sourceDriver, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("iofs: %w", err)
	}

	defer func() {
		_ = sourceDriver.Close()
	}()

	dbConn := stdlib.OpenDB(*conn.Config())
	defer func() {
		_ = dbConn.Close()
	}()

	// Create a new pgx migration driver instance.
	dbDriver, err := pgxmigrate.WithInstance(dbConn, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("pgxmigrate: %w", err)
	}

	// Create the migrate instance using the source and database drivers.
	m, err := migrate.NewWithInstance(
		"iofs", sourceDriver,
		"pgx", dbDriver,
	)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	// Run the up migrations.
	if err = m.Up(); err != nil && !eris.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("up: %w", err)
	}

	return nil
}
