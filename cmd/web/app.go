package main

import (
	"database/sql"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type App struct {
	DB        *sql.DB
	Session   *scs.SessionManager
	InfoLog   *log.Logger
	ErrorLog  *log.Logger
	WaitGroup *sync.WaitGroup
}

func initApp() *App {
	return &App{
		DB:        initDB(),
		Session:   initSession(),
		InfoLog:   log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:  log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		WaitGroup: &sync.WaitGroup{},
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
