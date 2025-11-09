package postgresrepo

import (
	"context"
	"fmt"
	"time"

	"database/sql"
	"gocloud.dev/postgres"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Init(user, password, dbname, host, port string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	r.db, err = postgres.Open(ctx, fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname))
	if err != nil {
		return err
	}

	if err := createTables(r.db); err != nil {
		return err
	}

	if _, err := r.db.ExecContext(ctx, "SET client_encoding TO 'UTF8'"); err != nil {
		_ = r.db.Close()
		return fmt.Errorf("set client_encoding: %w", err)
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
		id    TEXT PRIMARY KEY,
		name  TEXT      NOT NULL,
		ctype SMALLINT  NOT NULL CHECK (ctype IN (0, 1))
	);

	CREATE TABLE IF NOT EXISTS bank_accounts (
		id      TEXT PRIMARY KEY,
		name    TEXT   NOT NULL,
		balance DOUBLE PRECISION NOT NULL
	);


	CREATE TABLE IF NOT EXISTS operations (
		id            TEXT PRIMARY KEY,
		op_type       SMALLINT NOT NULL CHECK (op_type IN (0, 1)),
		account_id    TEXT     NOT NULL,
		category_id   TEXT     NOT NULL,
		amount        DOUBLE PRECISION NOT NULL,
		"timestamp"   TIMESTAMPTZ NOT NULL,  
		description   TEXT NOT NULL DEFAULT '',

		CONSTRAINT fk_operations_account
		FOREIGN KEY (account_id)  REFERENCES bank_accounts(id) ON DELETE CASCADE,

		CONSTRAINT fk_operations_category
		FOREIGN KEY (category_id) REFERENCES categories(id)   ON DELETE RESTRICT
	);

	`
	_, err := db.ExecContext(ctx, schema)
	return err
}

func (r *PostgresRepo) DB() *sql.DB {
	return r.db
}
