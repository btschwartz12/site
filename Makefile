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

# rust stuff
rust-server:
	cd rust && cargo build --release

run-rust-server: rust-server
	cd rust && godotenv -f ../.env ./target/release/rust-app

# c stuff
c-server:
	cd c && clang -o server app.c

run-c-server: c-server
	cd c && godotenv -f ../.env ./server