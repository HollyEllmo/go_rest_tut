.PHONY: build test run clean backup restore restore-testdata test-setup test-concurrent test-race

build:
	@mkdir -p bin
	@go build -o bin/go_rest_tut cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/go_rest_tut

clean:
	@rm -rf bin/

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations ${filter-out $@,$(MAKECMDGOALS)}

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down

# Database backup and restore commands
backup:
	@./scripts/backup_db.sh $(filter-out $@,$(MAKECMDGOALS))

restore:
	@./scripts/restore_testdata.sh $(filter-out $@,$(MAKECMDGOALS))

restore-testdata:
	@./scripts/restore_testdata.sh

# Testing commands
test-setup:
	@./scripts/setup_test_db.sh

test-concurrent:
	@go test ./cmd/service/inventory -v

test-race:
	@go test ./cmd/service/inventory -v -race