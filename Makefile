.PHONY: tidy
tidy:
	go mod tidy

.PHONY: run
run:
	go run main.go

.PHONY: build
build:
	go build -o bin/auction-service main.go

.PHONY: run.dev
run.dev:
	air

.PHONY: migration
migration:
	migrate create -ext sql -dir migration -seq ${name}

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go

.PHONY: seed
seed:
	go run cmd/seed/main.go

.PHONY: generate-swagger
generate-swagger:
	swag fmt -d main.go,./delivery/api && \
	swag init --parseDependency -d ./,./delivery/api --outputTypes json,yaml,go

.PHONY: test.cleancache
test.cleancache:
	go clean -testcache

.PHONY: test.unit
test.unit: test.cleancache
	go test -race ./...

.PHONY: test.cover
test.cover: test.cleancache
	go test -v -race ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func coverage.out