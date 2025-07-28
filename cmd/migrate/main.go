package main

import (
	"log"
	"os"

	"github.com/HollyEllmo/go_rest_tut/cmd/config"
	"github.com/HollyEllmo/go_rest_tut/cmd/db"
	mysqlCfg "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	db, err := db.NewMySQLStorage(mysqlCfg.Config{
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

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatal("Failed to create database driver:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://cmd/migrate/migrations",
		"mysql",
		driver,
	)
	if err != nil {
		log.Fatal("Failed to create migration instance:", err)
	}
	cmd := os.Args[(len(os.Args) - 1)]

	if cmd == "up" {
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to apply migrations:", err)
		} else {
			log.Println("Migrations applied successfully")
		}
	}

	if cmd == "down" {
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to revert migrations:", err)
		} else {
			log.Println("Migrations reverted successfully")
		}
	}

	if cmd == "force" {
		if len(os.Args) < 3 {
			log.Fatal("Usage: go run cmd/migrate/main.go force <version>")
		}
		version := os.Args[len(os.Args)-2]
		var v int
		if version == "none" {
			v = -1
		} else {
			// Simple conversion for version number
			switch version {
			case "20250725200923":
				v = 20250725200923
			case "20250725201722":
				v = 20250725201722
			case "20250725201739":
				v = 20250725201739
			case "20250725201812":
				v = 20250725201812
			case "20250726103947":
				v = 20250726103947
			case "20250726103948":
				v = 20250726103948
			default:
				log.Fatal("Unknown version:", version)
			}
		}
		if err := m.Force(v); err != nil {
			log.Fatal("Failed to force version:", err)
		} else {
			log.Printf("Forced database version to %d", v)
		}
	}
}
