package main

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func newMySQLConfig() *mysql.Config {
	return &mysql.Config{
		User:   getEnv("MYSQL_ROOT_USER", "root"),
		Passwd: getEnv("MYSQL_ROOT_PASSWORD", "password"),
		Net:    "tcp",
		DBName: getEnv("MYSQL_DATABASE", "todo"),
		Addr:   net.JoinHostPort(getEnv("MYSQL_HOST", "localhost"), getEnv("MYSQL_PORT", "3306")),
	}
}

func Test_TodoRepository(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	mysqlRepo, err := NewMySQLRepository(ctx, newMySQLConfig())
	require.NoError(t, err)

	testCases := []struct {
		name string
		repo TodoRepository
	}{
		{
			name: "InMemoryRepository",
			repo: NewInMemoryRepository(),
		},
		{
			name: "MySQLRepository",
			repo: mysqlRepo,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			t.Cleanup(cancel)
			err := tc.repo.Init(ctx)
			require.NoError(t, err)

			// Create
			err = tc.repo.Create(ctx, "task")
			require.NoError(t, err)

			// List
			todos, err := tc.repo.List(ctx)
			require.NoError(t, err)
			expectedTodos := []Todo{
				{
					ID:   1,
					Task: "task",
				},
			}
			require.Equal(t, expectedTodos, todos)

			// Done
			err = tc.repo.Done(ctx, 1)
			require.NoError(t, err)

			// Done twice to test ErrNotFound
			err = tc.repo.Done(ctx, 1)
			assert.ErrorIs(t, err, ErrNotFound)

			// List
			todos, err = tc.repo.List(ctx)
			require.NoError(t, err)
			var emptyTodos []Todo
			assert.Equal(t, emptyTodos, todos)
		})
	}
}
