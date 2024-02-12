# go-sample-todo

This is a simple todo list application written in Go.

## Getting Started

### Prerequisites

- Go 1.22 or later

### Installation

```sh
go get github.com/pddg/go-sample-todo
```

### Usage

Running in-memory mode. This mode does not require any database.

```sh
./go-sample-todo -in-memory
```

Running with MySQL.

```sh
./go-sample-todo -mysql-host=localhost -mysql-port=3306 -mysql-user=root -mysql-password=root
```

## Endpoints

### `GET /todo`

Get all todos.

```sh
curl -X GET http://localhost:8080/todo
```

### `POST /todo`

Create a new todo.

```sh
curl -X POST -H "Content-Type: application/json" -d '{"task": "Buy milk"}' http://localhost:8080/todo
```

### `DELETE /todo/:id`

Delete a todo.

```sh
curl -X DELETE http://localhost:8080/todo/1
```

## Author

- pddg
