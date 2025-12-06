set shell := ["bash", "-eu", "-o", "pipefail", "-c"]

pg_port := "5544"
redis_port := "6381"
meili_port := "7701"

dev-up:
	./scripts/dev/services.sh up

dev-down:
	./scripts/dev/services.sh down

dev-status:
	./scripts/dev/services.sh status

dev-db-setup:
	@echo "Setting up database role and database..."
	@PG_SUPERUSER="$${PGUSER:-$${USER:-postgres}}" && \
	psql -h 127.0.0.1 -p {{pg_port}} -U "$$PG_SUPERUSER" -d postgres -c "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'mgsearch') THEN CREATE ROLE mgsearch LOGIN PASSWORD 'mgsearch'; END IF; END \$\$;" && \
	psql -h 127.0.0.1 -p {{pg_port}} -U "$$PG_SUPERUSER" -d postgres -c "DO \$\$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'mgsearch') THEN CREATE DATABASE mgsearch OWNER mgsearch; END IF; END \$\$;" && \
	echo "Database setup complete!"

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

