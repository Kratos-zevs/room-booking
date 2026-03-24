package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func New() *sql.DB {
	connStr := "postgres://postgres:postgres@db:5432/booking?sslmode=disable"

	var database *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		database, err = sql.Open("postgres", connStr)
		if err == nil {
			err = database.Ping()
			if err == nil {
				log.Println("DB connected")
				return database
			}
		}

		log.Println("waiting for DB...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("failed to connect to DB:", err)
	return nil
}