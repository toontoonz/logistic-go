.PHONY: test up

test:
	cd backend && go test ./...

up:
	docker compose up --build
