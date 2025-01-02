include .dev.env
export

.PHONY: setup dev certs db_up db_down test_db gen_templ gen_sqlc gen migrate_up migrate_down migrate_create

setup:
	go mod tidy
	go mod download
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	brew install mkcert

dev: db_up
	DOTENV=.dev.env air

certs:
	mkdir -p certs
	cd certs && mkcert -install
	cd certs && mkcert localhost "*.localhost"

db_up:
	docker compose up db -d

db_down:
	docker compose down db

test_db:
	docker compose up test_db

gen_templ:
	templ generate

gen_sqlc:
	sqlc generate -f internal/db/sqlc.yaml

gen: gen_templ gen_sqlc

migrate_up:
	migrate -database "${DATABASE_URL}" -path internal/db/migrations up

migrate_down:
	migrate -database "${DATABASE_URL}" -path internal/db/migrations down

migrate_create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)
