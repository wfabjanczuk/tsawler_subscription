package main

import (
	"context"
	"encoding/gob"
	"github.com/alexedwards/scs/v2"
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var testApp App

func TestMain(m *testing.M) {
	gob.Register(models.User{})

	pathToTemplates = "./templates"
	pathToManual = "./../../pdf"
	pathToTmp = "./../../tmp"

	session := scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = true

	testApp = App{
		DB:            nil,
		Session:       session,
		Models:        models.TestNew(nil),
		InfoLog:       log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime),
		ErrorLog:      log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
		WaitGroup:     &sync.WaitGroup{},
		ErrorChan:     make(chan error),
		ErrorChanDone: make(chan bool),
	}

	mailerErrorChan := make(chan error)
	mailerChan := make(chan Message, 100)
	mailerDoneChan := make(chan bool)

	testApp.Mailer = &Mail{
		WaitGroup:  testApp.WaitGroup,
		ErrorChan:  mailerErrorChan,
		MailerChan: mailerChan,
		DoneChan:   mailerDoneChan,
	}

	go func() {
		for {
			select {
			case <-testApp.Mailer.MailerChan:
				testApp.WaitGroup.Done()
			case <-testApp.Mailer.ErrorChan:
				testApp.WaitGroup.Done()
			case <-testApp.Mailer.DoneChan:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case err := <-testApp.ErrorChan:
				testApp.ErrorLog.Println(err)
			case <-testApp.ErrorChanDone:
				return
			}
		}
	}()

	os.Exit(m.Run())
}

func getCtx(req *http.Request) context.Context {
	ctx, err := testApp.Session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}

	return ctx
}
