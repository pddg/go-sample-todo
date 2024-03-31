package main

import (
	"context"
	"sync"
)

type InMemoryRepository struct {
	mutex    sync.Mutex
	todos    []Todo
	latestID uint64
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		todos: make([]Todo, 0),
	}
}

func (i *InMemoryRepository) List(ctx context.Context) ([]Todo, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if len(i.todos) == 0 {
		var emptyTodos []Todo
		return emptyTodos, nil
	}
	copiedTodos := make([]Todo, len(i.todos))
	copy(copiedTodos, i.todos)
	return copiedTodos, nil
}

func (i *InMemoryRepository) Create(ctx context.Context, task string) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.latestID++
	i.todos = append(i.todos, Todo{
		ID:   i.latestID,
		Task: task,
	})
	return nil
}

func (i *InMemoryRepository) Done(ctx context.Context, id uint64) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	for idx, todo := range i.todos {
		if id == todo.ID {
			i.todos = append(i.todos[:idx], i.todos[idx+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (i *InMemoryRepository) Init(ctx context.Context) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.todos = make([]Todo, 0)
	i.latestID = 0
	return nil
}
