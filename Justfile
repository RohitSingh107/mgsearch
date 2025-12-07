set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

mongodb_port := "27017"
redis_port := "6381"
meili_port := "7701"

dev-up:
	./scripts/dev/services.sh up

dev-down:
	./scripts/dev/services.sh down

dev-status:
	./scripts/dev/services.sh status

dev-db-setup:
	@echo "MongoDB database setup is automatic - collections and indexes are created on first run"
	@echo "Database will be created automatically when the application connects"

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

