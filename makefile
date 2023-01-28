run:
	go run main.go
sqlc:
	sqlc generate
compile:
	sqlc compile

postgres:
	psql saasita

docker:
	docker compose up