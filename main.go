package main

import (
	"database/sql"
	"log"
	"simplebank/api"
	db "simplebank/db/sqlc"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:secret@localhost:3333/simple_bank?sslmode=disable"
	address  = "0.0.0.0:8088"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)

	if err != nil {
		log.Fatal(err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	log.Fatal(server.Start(address))

}
