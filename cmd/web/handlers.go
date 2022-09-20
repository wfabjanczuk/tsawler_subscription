package main

import "net/http"

func (a *App) HomePage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "home.page.gohtml", nil)
}

func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "login.page.gohtml", nil)
}

func (a *App) PostLoginPage(w http.ResponseWriter, r *http.Request) {
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
}

func (a *App) RegisterPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register.page.gohtml", nil)
}

func (a *App) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
}

func (a *App) ActivateAccount(w http.ResponseWriter, r *http.Request) {
}
