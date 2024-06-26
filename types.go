package main

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

type Todo struct {
	ID   uint64 `json:"id" db:"id"`
	Task string `json:"task" db:"task"`
}
