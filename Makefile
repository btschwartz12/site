server:
	CGO_ENABLED=0 go build -o server main.go

swagger:
	swag init --output api/swagger -g api/swagger/main.go

sqlc:
	cd internal/repo/db && sqlc generate

run-server: swagger sqlc server 
	godotenv -f .env ./server --port 8080 --dev

clean:
	rm -f server