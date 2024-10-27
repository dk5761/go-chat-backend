.PHONY: build run migrate

build:
	go build -o server ./cmd/server

run: build
	./server

migrate:
	migrate -path ./migrations -database "postgres://youruser:yourpassword@localhost:5432/yourdb?sslmode=disable" up

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

test:
	go test ./...
