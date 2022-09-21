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
	mux.Get("/activate", a.ActivateAccount)

	mux.Mount("/members", a.authRouter())

	return mux
}

func (a *App) authRouter() http.Handler {
	mux := chi.NewRouter()
	mux.Use(a.Auth)

	mux.Get("/plans", a.ChooseSubscription)
	mux.Get("/subscribe", a.SubscribeToPlan)

	return mux
}

func (a *App) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.Session.Exists(r.Context(), "userID") {
			a.Session.Put(r.Context(), "warning", "You must log in first!")
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}
		next.ServeHTTP(w, r)
	})
}
