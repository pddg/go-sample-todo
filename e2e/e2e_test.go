package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name string
		port string
	}{
		{
			name: "InMemory mode",
			port: inMemoryPort,
		},
		{
			name: "MySQL mode",
			port: mysqlPort,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			runTodoServer(t, tc.port)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			t.Cleanup(cancel)

			// Setup
			client := NewAPIClient(tc.port)
			err := client.Init(ctx)
			require.NoError(t, err)

			// Create
			resp, err := client.CreateTask(ctx, "task")
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusCreated, resp.StatusCode)

			// List
			resp, err = client.ListTasks(ctx)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.JSONEq(t, `{"todos":[{"id":1,"task":"task"}]}`, string(body))

			// Done
			resp, err = client.DoneTask(ctx, 1)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// Done twice to test NotFound
			resp, err = client.DoneTask(ctx, 1)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}

const (
	inMemoryPort = "8080"
	mysqlPort    = "8081"
)

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type APIClient struct {
	port string
}

func NewAPIClient(port string) *APIClient {
	return &APIClient{port: port}
}

func (c *APIClient) do(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, "http://localhost:"+c.port+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	return http.DefaultClient.Do(req)
}

func (c *APIClient) CreateTask(ctx context.Context, task string) (*http.Response, error) {
	return c.do(ctx, "POST", "/todo", bytes.NewBufferString(fmt.Sprintf(`{"task":"%s"}`, task)))
}

func (c *APIClient) ListTasks(ctx context.Context) (*http.Response, error) {
	return c.do(ctx, "GET", "/todo", nil)
}

func (c *APIClient) DoneTask(ctx context.Context, id uint64) (*http.Response, error) {
	return c.do(ctx, "DELETE", fmt.Sprintf("/todo/%d", id), nil)
}

func (c *APIClient) Init(ctx context.Context) error {
	resp, err := c.do(ctx, "POST", "/initialize", nil)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to initialize: %s", resp.Status)
	}
	return nil
}

func runTodoServer(t *testing.T, port string) {
	bin := "../build/sample-todo"
	args := []string{"-port", port}
	switch port {
	case inMemoryPort:
		args = append(args, "-in-memory")
	case mysqlPort:
		args = append(args,
			"-mysql-host", getEnv("MYSQL_HOST", "localhost"),
			"-mysql-port", getEnv("MYSQL_PORT", "3306"),
			"-mysql-user", getEnv("MYSQL_USER", "todo"),
			"-mysql-password", getEnv("MYSQL_PASSWORD", "todo"),
		)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	srv := exec.CommandContext(ctx, bin, args...)
	srv.WaitDelay = 100 * time.Millisecond
	srv.Cancel = func() error {
		return srv.Process.Signal(syscall.SIGTERM)
	}
	go func() {
		if err := srv.Run(); err != nil {
			t.Log(err)
		}
	}()

	healthCheckCtx, healthCheckCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer healthCheckCancel()
	for {
		select {
		case <-healthCheckCtx.Done():
			t.Fatalf("server is not healthy in time")
		default:
			resp, err := http.Get("http://localhost:" + port + "/healthz")
			if err == nil && resp.StatusCode == http.StatusOK {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
