MYSQL_PORT?=33060
MYSQL_ROOT_USER?=root
MYSQL_ROOT_PASSWORD?=root

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -o sample-todo .

.PHONY: test
test:
	MYSQL_PORT=$(MYSQL_PORT) MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) go test -race .

.PHONY: test-e2e
test-e2e: build
	MYSQL_PORT=$(MYSQL_PORT) MYSQL_ROOT_PASSWORD=$(MYSQL_ROOT_PASSWORD) go test ./e2e

.PHONY: start-mysql
start-mysql:
	docker run -d --rm \
		--name todo-mysql \
		-e MYSQL_ROOT_PASSWORD=root \
		-e MYSQL_DATABASE=todo \
		-e MYSQL_USER=todo \
		-e MYSQL_PASSWORD=todo \
		-p $(MYSQL_PORT):3306 \
		mysql:8.0

.PHONY: stop-mysql
stop-mysql:
	docker stop todo-mysql
