server:
	CGO_ENABLED=0 go build -o server main.go

rust-server:
	cd rust && cargo build --release

swagger:
	swag init --output api/swagger -g api/swagger/main.go

sqlc:
	cd internal/repo/db && sqlc generate

run-server: swagger sqlc server 
	godotenv -f .env ./server --port 8080 --dev

run-rust-server: rust-server
	cd rust && godotenv -f ../.env ./target/release/rust-app

clean:
	rm -f server

