package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type WriteInterceptor struct {
	w      http.ResponseWriter
	status int
}

func (w *WriteInterceptor) Header() http.Header {
	return w.w.Header()
}

func (w *WriteInterceptor) Write(body []byte) (int, error) {
	return w.w.Write(body)
}
func (w *WriteInterceptor) WriteHeader(statusCode int) {
	w.status = statusCode
	w.w.WriteHeader(statusCode)
}

func AccessLogMiddleware(next http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		interceptor := &WriteInterceptor{w: w}
		next.ServeHTTP(interceptor, r)
		responseTime := time.Now().Sub(startTime)
		logger.Info("access log", "method", r.Method, "path", r.URL.Path, "status", interceptor.status, "response_time", responseTime)
	})
}
