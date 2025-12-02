.PHONY: up down build run docker-build logs

# Docker Compose commands
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Go commands
build:
	go build -o bin/server ./cmd/api

run:
	go run ./cmd/api/main.go

# Docker build command
docker-build:
	docker build -t keepa-backend .
