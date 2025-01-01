.PHONY: dev certs

dev:
	air

certs:
	mkdir -p certs
	cd certs && mkcert -install
	cd certs && mkcert localhost "*.localhost"
