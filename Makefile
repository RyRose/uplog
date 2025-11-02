.PHONY: all
all: build

.PHONY: build
build: types
	tailwindcss -i web/app/input.css -o web/static/css/output.css & sqlc generate & templ generate & wait
	go build -o=./tmp/uplog ./cmd/uplog

.PHONY: run
run: build
	tmp/uplog

.PHONY: gstatus
gstatus:
	go run ./cmd/goose status

.PHONY: gup
gup:
	go run ./cmd/goose up

.PHONY: gdown
gdown:
	go run ./cmd/goose down

.PHONY: serve
serve:
	./scripts/air.sh

.PHONY: test
test: build
	go test ./...
	busted

.PHONY: clean
clean:
	./scripts/clean.sh

.PHONY: tidy
tidy:
	go mod tidy -v

.PHONY: nix
nix:
	nix develop

.PHONY: deps
deps:
	nix flake update
	npm install
	go get -u ./...

.PHONY: format
format:
	go fmt ./...
	stylua ./config
	swag fmt

.PHONY: docs
docs:
	swag init -g ./cmd/uplog/main.go

.PHONY: types
types:
	go run ./cmd/generatetypes > ./config/lib/typedefinitions.lua

.PHONY: docker
docker:
	docker build .

.PHONY: lint
lint:
	golangci-lint run

install:
	npm ci

.PHONY: gha
gha: install build test lint

.PHONY: ci
ci: clean docs format gha docker

