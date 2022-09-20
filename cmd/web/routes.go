package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (a *App) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer, a.Session.LoadAndSave)

	mux.Get("/", a.HomePage)
	mux.Get("/login", a.LoginPage)
	mux.Post("/login", a.PostLoginPage)
	mux.Get("/logout", a.Logout)
	mux.Get("/register", a.RegisterPage)
	mux.Post("/register", a.PostRegisterPage)
	mux.Get("/activate-account", a.ActivateAccount)

	return mux
}
