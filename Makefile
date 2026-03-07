postgres:
	docker run --name postgres -p 3333:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:18

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:3333/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:3333/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

server:
	go run main.go

mock:
	mockgen -package mockdb --destination db/mock/store.go simplebank/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock