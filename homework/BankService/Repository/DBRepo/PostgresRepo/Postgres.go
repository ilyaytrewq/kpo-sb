package postgresrepo

import (
	"context"
	"fmt"
	"time"

	"gocloud.dev/postgres"
	"database/sql"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Init(user string, password string, dbname string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()	

	var err error
	r.db, err = postgres.Open(ctx, fmt.Sprintf("postgres://%s:%s@localhost:5432/%s?sslmode=disable", user, password, dbname))
	if err != nil {
		return err
	}

	if err := createTables(r.db); err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepo) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

func createTables(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	schema := `
	CREATE TABLE IF NOT EXISTS categories (
		id UNIQUE TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type INT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS bank_accounts (
		id UNIQUE TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		balance FLOAT8 NOT NULL
	);
	CREATE TABLE IF NOT EXISTS operations (
		id UNIQUE TEXT PRIMARY KEY,
		account_id TEXT NOT NULL,
		category_id TEXT NOT NULL,
		amount FLOAT8 NOT NULL,
		timestamp TIMESTAMPTZ NOT NULL,
		FOREIGN KEY (account_id) REFERENCES bank_accounts(id),
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);
	`
	_, err := db.ExecContext(ctx, schema)
	return err
}