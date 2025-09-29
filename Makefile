.PHONY: all
all: build

.PHONY: build
build:
	tailwindcss -i web/app/input.css -o web/static/css/output.css
	sqlc generate
	templ generate
	go build -o=./tmp/uplog ./cmd/uplog

.PHONY: run
run: build
	PORT=8080 tmp/uplog

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
test:
	go test ./...

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
	npm install

.PHONY: format
format:
	swag fmt

.PHONY: docs
docs:
	swag init

