.PHONY: test up

test:
	cd backend && go test ./...

up:
	docker compose up --build

# Publish Postgres on 127.0.0.1:5433 and Redis on 127.0.0.1:6380
# so `go run` on the host can reach them. See the README.
up-ports:
	docker compose -f docker-compose.yml -f docker-compose.ports.yml up --build
