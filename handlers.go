package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type TodoRepository interface {
	List(ctx context.Context) ([]Todo, error)
	Create(ctx context.Context, task string) error
	Done(ctx context.Context, id uint64) error
}

func NewListTodosHandler(repo TodoRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		todos, err := repo.List(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		bodyBytes, err := json.Marshal(
			map[string][]Todo{
				"todos": todos,
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bodyBytes)
	}
}

func NewCreateTodoHandler(repo TodoRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type createTodoRequest struct {
			Task string `json:"task"`
		}
		var todo createTodoRequest
		if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if err := repo.Create(r.Context(), todo.Task); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func NewDoneTodoHandler(repo TodoRepository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if err := repo.Done(r.Context(), id); err != nil {
			if errors.Is(err, ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(err.Error()))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
