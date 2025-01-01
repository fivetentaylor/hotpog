.PHONY: dev certs

setup:
	go mod tidy
	go mod download
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	brew install mkcert

dev: db
	air

certs:
	mkdir -p certs
	cd certs && mkcert -install
	cd certs && mkcert localhost "*.localhost"

db:
	docker compose up db

test_db:
	docker compose up test_db
