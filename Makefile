.PHONY: dev certs

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
