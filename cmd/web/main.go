package main

import (
	"fmt"
	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	"net/http"
)

const port = "80"

func main() {
	app := initApp()
	app.serve()
}

func (a *App) serve() {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: a.routes(),
	}

	a.InfoLog.Println("Starting web server...")
	err := server.ListenAndServe()

	if err != nil {
		a.ErrorLog.Panic(err)
	}
}
