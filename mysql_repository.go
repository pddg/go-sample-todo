package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type MySQLRepository struct {
	db *sqlx.DB
}

func NewMySQLRepository(ctx context.Context, conf *mysql.Config) (*MySQLRepository, error) {
	for {
		db, err := sqlx.Open("mysql", conf.FormatDSN())
		if err == nil {
			err = db.Ping()
			if err == nil {
				return &MySQLRepository{db: db}, nil
			}
		}
		slog.Error("Failed to connect to MySQL", "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(1 * time.Second):
		}
	}
}

func (m *MySQLRepository) List(ctx context.Context) ([]Todo, error) {
	var todos []Todo
	if err := m.db.SelectContext(ctx, &todos, "SELECT * FROM todos"); err != nil {
		return nil, fmt.Errorf("failed to select todos: %w", err)
	}
	return todos, nil
}

func (m *MySQLRepository) Create(ctx context.Context, task string) error {
	if _, err := m.db.ExecContext(ctx, "INSERT INTO todos (task) VALUES (?)", task); err != nil {
		return fmt.Errorf("failed to insert todo: %w", err)
	}
	return nil
}

func (m *MySQLRepository) Done(ctx context.Context, id uint64) error {
	res, err := m.db.ExecContext(ctx, "DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}
	deleted, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if deleted == 0 {
		return ErrNotFound
	}
	return nil
}

func (m *MySQLRepository) Init(ctx context.Context) error {
	_, err := m.db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS todos (id INT AUTO_INCREMENT PRIMARY KEY, task TEXT)")
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	if _, err := m.db.ExecContext(ctx, "TRUNCATE TABLE todos"); err != nil {
		return fmt.Errorf("failed to truncate table: %w", err)
	}
	return nil
}
