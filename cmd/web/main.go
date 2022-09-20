package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const port = "80"

func main() {
	db := initDB()

	if err := db.Ping(); err != nil {
		log.Println(err)
	}
}

func initDB() *sql.DB {
	conn := connectToDB()
	if conn == nil {
		log.Panicln("cannot connect to database")
	}

	return conn
}

func connectToDB() *sql.DB {
	dsn := os.Getenv("DSN")
	retryingDelay := 1 * time.Second
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		connection, err := openDB(dsn)
		if err != nil {
			log.Printf("retry #%d: postgres not yet ready...", i)
		} else {
			log.Printf("retry #%d: connected to database", i)
			return connection
		}

		time.Sleep(retryingDelay)
	}

	return nil
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
