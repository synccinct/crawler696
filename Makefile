run:
	go run ./cmd/main.go

dev:
	docker compose up --build --remove-orphans

test:
	go test ./... -v

lint:
	golangci-lint run

prod-build:
	docker compose --profile prod build
