MYSQL_PORT?=33060
MYSQL_ROOT_USER?=root
MYSQL_ROOT_PASSWORD?=root
VERSION:=$(shell cat VERSION)

.PHONY: build
build:
	CGO_ENABLED=0 go build -trimpath -o build/sample-todo .

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

# === Debian package ===
DEB_PACKAGE_CONTENTS=/opt/go-sample-todo/bin/sample-todo $(subst debmeta,,$(wildcard debmeta/DEBIAN/*)) /etc/systemd/system/sample-todo.service
build/deb/%: debmeta/%
	mkdir -p $(dir $@)
	cp $< $@

build/deb/DEBIAN/control: debmeta/DEBIAN/control VERSION
	mkdir -p $(dir $@)
	cp $< $@
	sed -i 's/\%VERSION\%/$(VERSION)/' $@

build/deb/opt/go-sample-todo/bin/sample-todo: build
	mkdir -p $(dir $@)
	cp build/sample-todo $@

build/go-sample-todo_$(VERSION)_amd64.deb: $(addprefix build/deb,$(DEB_PACKAGE_CONTENTS))
	fakeroot dpkg-deb --build build/deb $(dir $@)

.PHONY: package
package: build/go-sample-todo_$(VERSION)_amd64.deb
# ======

.PHONY: clean
clean:
	rm -rf build
