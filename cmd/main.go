package main

import (
	"database/sql"
	"log"

	"github.com/HollyEllmo/go_rest_tut/cmd/api"
	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/db"
	"github.com/go-sql-driver/mysql"
)

func main() {
	db, err := db.NewMySQLStorage(mysql.Config{
		User:                 config.Envs.DBUser,
		Passwd:               config.Envs.DBPassword,
		Addr:                 config.Envs.DBAddress,
		DBName:               config.Envs.DBName,
		Net:                  "tcp",
		AllowNativePasswords: true,
		ParseTime:            true,
	})

	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	initStorage(db)

	server := api.NewAPIServer("localhost:8080", db)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}

func initStorage(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("DB: Successfully connected!")
}
