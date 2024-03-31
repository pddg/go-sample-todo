package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-sql-driver/mysql"
)

func runServer(config *mysql.Config, port uint) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var (
		err  error
		repo TodoRepository
	)
	if config == nil {
		repo = NewInMemoryRepository()
	} else {
		repo, err = NewMySQLRepository(ctx, config)
		if err != nil {
			return err
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /todo", NewListTodosHandler(repo))
	mux.HandleFunc("POST /todo", NewCreateTodoHandler(repo))
	mux.HandleFunc("DELETE /todo/{id}", NewDoneTodoHandler(repo))
	mux.HandleFunc("POST /initialize", NewInitHandler(repo))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: AccessLogMiddleware(mux),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		slog.Info("Shutting down server")
		if err := server.Shutdown(context.WithoutCancel(ctx)); err != nil {
			slog.Error("Server shutdown failed", "error", err)
		}
		slog.Info("Server has been shut down")
	}()

	slog.Info("Starting the server", "port", port)
	if err = server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		} else {
			slog.Error("Server stopped", "error", err)
		}
	}
	wg.Wait()
	return err
}

func main() {
	var (
		port          uint
		inMemory      bool
		mysqlHost     string
		mysqlPort     uint
		mysqlUser     string
		mysqlPassword string
	)
	flag.UintVar(&port, "port", 8080, "Port to listen on")
	flag.BoolVar(&inMemory, "in-memory", false, "Use in-memory repository")
	flag.StringVar(&mysqlHost, "mysql-host", "localhost", "MySQL host")
	flag.UintVar(&mysqlPort, "mysql-port", 3306, "MySQL port")
	flag.StringVar(&mysqlUser, "mysql-user", "root", "MySQL user")
	flag.StringVar(&mysqlPassword, "mysql-password", "", "MySQL password")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	var conf *mysql.Config
	if !inMemory {
		conf = mysql.NewConfig()
		conf.Net = "tcp"
		conf.Addr = net.JoinHostPort(mysqlHost, fmt.Sprintf("%d", mysqlPort))
		conf.User = mysqlUser
		conf.Passwd = mysqlPassword
		conf.DBName = "todo"
	}

	if err := runServer(conf, port); err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Info("Server stopped")
		} else {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}
}
