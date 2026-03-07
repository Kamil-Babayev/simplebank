package main

import (
	"database/sql"
	"log"
	"simplebank/api"
	db "simplebank/db/sqlc"
	"simplebank/util"

	_ "github.com/lib/pq"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)

	if err != nil {
		log.Fatal("Failed to connect to data base: ", err)
	}

	store := db.NewStore(conn)
	server := api.NewServer(store)

	log.Fatal(server.Start(config.ServerAddress))

}
