include .dev.env
export

.PHONY: setup dev certs dev_up dev_down test_up test_down gen_templ gen_sqlc gen_tailwind gen migrate_up migrate_down migrate_create

setup:
	go mod tidy
	go mod download
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	brew install mkcert

dev: dev_up
	DOTENV=.dev.env air

certs:
	mkdir -p certs
	cd certs && mkcert -install
	cd certs && mkcert localhost "*.localhost"

dev_up:
	docker compose up -d

dev_down:
	docker compose down

test_up:
	docker compose -f docker-compose.test.yml up -d

test_down:
	docker compose -f docker-compose.test.yml down

gen_templ:
	templ generate

gen_sqlc:
	sqlc generate -f internal/db/sqlc.yaml

gen_tailwind:
	npx tailwindcss -i ./internal/router/static/input.css -o ./internal/router/static/output.css

gen: gen_templ gen_sqlc gen_tailwind

migrate_up:
	migrate -database "${DATABASE_URL}" -path internal/db/migrations up

migrate_down:
	migrate -database "${DATABASE_URL}" -path internal/db/migrations down

migrate_create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)
