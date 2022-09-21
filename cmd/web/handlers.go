package main

import "net/http"

func (a *App) HomePage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "home.page.gohtml", nil)
}

func (a *App) LoginPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "login.page.gohtml", nil)
}

func (a *App) PostLoginPage(w http.ResponseWriter, r *http.Request) {
	err := a.Session.RenewToken(r.Context())
	if err != nil {
		a.ErrorLog.Println(err)
	}

	err = r.ParseForm()
	if err != nil {
		a.ErrorLog.Println(err)
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := a.Models.User.GetByEmail(email)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	validPassword, err := user.PasswordMatches(password)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if !validPassword {
		a.sendEmail(Message{
			ToAddress: email,
			Subject:   "Failed login attempt",
			Data:      "Invalid login attempt!",
		})

		a.Session.Put(r.Context(), "error", "Invalid credentials.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	a.Session.Put(r.Context(), "userID", user.ID)
	a.Session.Put(r.Context(), "user", user)
	a.Session.Put(r.Context(), "flash", "Successful login!")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	err := a.Session.Destroy(r.Context())
	if err != nil {
		a.ErrorLog.Println(err)
	}

	err = a.Session.RenewToken(r.Context())
	if err != nil {
		a.ErrorLog.Println(err)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) RegisterPage(w http.ResponseWriter, r *http.Request) {
	a.render(w, r, "register.page.gohtml", nil)
}

func (a *App) PostRegisterPage(w http.ResponseWriter, r *http.Request) {
}

func (a *App) ActivateAccount(w http.ResponseWriter, r *http.Request) {
}
