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

# make migration name=create_users_table
.PHONY: migration
migration:
	migrate create -ext sql -dir migration -seq ${name}

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go

# make migrate-rollback steps=1
.PHONY: migrate-rollback
migrate-rollback:
	go run cmd/migrate/main.go -rollback -steps=$(steps)

.PHONY: seed
seed:
	go run cmd/seed/main.go

.PHONY: generate-swagger
generate-swagger:
	go run ./tool/swag fmt -d main.go,./delivery/api && \
	go run ./tool/swag init --parseDependency -d ./,./delivery/api --outputTypes json,yaml,go

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