package main

import (
	"context"
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
	panic("not implemented") // TODO: Implement
}

func (m *MySQLRepository) Create(ctx context.Context, task string) error {
	panic("not implemented") // TODO: Implement
}

func (m *MySQLRepository) Done(ctx context.Context, id uint64) error {
	panic("not implemented") // TODO: Implement
}
