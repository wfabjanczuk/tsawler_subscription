package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const port = "80"

type App struct {
	DB        *sql.DB
	Session   *scs.SessionManager
	Models    *models.Models
	InfoLog   *log.Logger
	ErrorLog  *log.Logger
	WaitGroup *sync.WaitGroup
	Mailer    *Mail
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

func initApp() *App {
	db := initDB()
	wg := &sync.WaitGroup{}

	return &App{
		DB:        db,
		Session:   initSession(),
		Models:    models.New(db),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		Mailer:    createMail(wg),
		WaitGroup: wg,
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
			log.Println(err)
			log.Printf("retry #%d: postgres not yet ready...\n", i)
		} else {
			log.Printf("retry #%d: connected to database\n", i)
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

func initSession() *scs.SessionManager {
	gob.Register(models.User{})

	session := scs.New()
	session.Store = redisstore.New(initRedis())
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	return session
}

func initRedis() *redis.Pool {
	return &redis.Pool{
		MaxIdle: 10,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", os.Getenv("REDIS"))
		},
	}
}

func createMail(wg *sync.WaitGroup) *Mail {
	errorChan := make(chan error)
	mailerChan := make(chan Message, 100)
	mailerDoneChan := make(chan bool)

	return &Mail{
		Domain:      "localhost",
		Host:        "localhost",
		Port:        1025,
		FromName:    "Info",
		FromAddress: "info@mycompany.com",
		Encryption:  "none",
		ErrorChan:   errorChan,
		MailerChan:  mailerChan,
		DoneChan:    mailerDoneChan,
		WaitGroup:   wg,
	}
}

func (a *App) listenForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.shutdown()
	os.Exit(0)
}

func (a *App) shutdown() {
	a.InfoLog.Println("running cleanup tasks...")
	a.WaitGroup.Wait()
	a.Mailer.DoneChan <- true

	close(a.Mailer.ErrorChan)
	close(a.Mailer.MailerChan)
	close(a.Mailer.DoneChan)

	a.InfoLog.Println("closing channels and shutting down...")
}
