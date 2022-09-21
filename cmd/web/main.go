package main

import (
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	app := initApp()
	go app.listenForMail()
	go app.listenForShutdown()
	app.serve()
}
