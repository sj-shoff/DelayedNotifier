.PHONY: run build migrate-up migrate-down docker-up docker-down curl-test
include .env
export

run:
	go run cmd/app/main.go

build:
	go build -o bin/delayed-notifier cmd/app/main.go

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

docker-clean:
	docker-compose down -v --remove-orphans

migrate-up:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down

curl-test:
	@echo "=== Testing Delayed Notifier API ==="
	
	@echo "1. Testing create notification (email channel)..."
	@curl -X POST http://localhost:${SERVER_PORT}/api/v1/notify \
		-H "Content-Type: application/json" \
		-d '{"user_id": "user@example.com", "channel": "email", "message": "Test message", "send_at": "2025-11-16T10:00:00Z"}' \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "2. Testing create notification (telegram channel)..."
	@curl -X POST http://localhost:${SERVER_PORT}/api/v1/notify \
		-H "Content-Type: application/json" \
		-d '{"user_id": "123456789", "channel": "telegram", "message": "Test telegram", "send_at": "2025-11-16T11:00:00Z"}' \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "3. Getting all notifications..."
	@curl -X GET http://localhost:${SERVER_PORT}/api/v1/notifications \
		-H "Content-Type: application/json" \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "4. Getting notification status (replace ID with actual from previous response)..."
	@curl -X GET http://localhost:${SERVER_PORT}/api/v1/notify/actual-id-here \
		-H "Content-Type: application/json" \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "5. Canceling notification (DELETE)..."
	@curl -X DELETE http://localhost:${SERVER_PORT}/api/v1/notify/actual-id-here \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "6. Testing invalid channel..."
	@curl -X POST http://localhost:${SERVER_PORT}/api/v1/notify \
		-H "Content-Type: application/json" \
		-d '{"user_id": "user@example.com", "channel": "invalid", "message": "Test", "send_at": "2025-11-16T10:00:00Z"}' \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "7. Testing past send_at..."
	@curl -X POST http://localhost:${SERVER_PORT}/api/v1/notify \
		-H "Content-Type: application/json" \
		-d '{"user_id": "user@example.com", "channel": "email", "message": "Test", "send_at": "2020-01-01T00:00:00Z"}' \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "8. Getting non-existent notification..."
	@curl -X GET http://localhost:${SERVER_PORT}/api/v1/notify/invalid-id \
		-w "\n=== Response: %{http_code}\n\n"
	
	@echo "=== Testing completed ==="