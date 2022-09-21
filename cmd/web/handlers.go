package main

import (
	"fmt"
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"html/template"
	"net/http"
)

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
	err := r.ParseForm()
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to create user.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	u := models.User{
		Email:     r.Form.Get("email"),
		FirstName: r.Form.Get("first-name"),
		LastName:  r.Form.Get("last-name"),
		Password:  r.Form.Get("password"),
		Active:    0,
		IsAdmin:   0,
	}

	_, err = a.Models.User.Insert(u)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to create user.")
		http.Redirect(w, r, "/register", http.StatusSeeOther)
		return
	}

	url := GenerateTokenFromString(fmt.Sprintf("http://localhost/activate?email=%s", u.Email))
	a.InfoLog.Println(url)

	msg := Message{
		ToAddress: u.Email,
		Subject:   "Activate your account",
		Template:  "confirmation-email",
		Data:      template.HTML(url),
	}

	a.sendEmail(msg)
	a.Session.Put(r.Context(), "flash", "Confirmation email sent.")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (a *App) ActivateAccount(w http.ResponseWriter, r *http.Request) {
	url := r.RequestURI
	testURL := fmt.Sprintf("http://localhost%s", url)

	if !VerifyToken(testURL) {
		a.Session.Put(r.Context(), "error", "Invalid token.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u, err := a.Models.User.GetByEmail(r.URL.Query().Get("email"))
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "No user found.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	u.Active = 1
	err = u.Update()
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to update user.")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	a.Session.Put(r.Context(), "flash", "Account activated - you can now log in.")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) ChooseSubscription(w http.ResponseWriter, r *http.Request) {
	if !a.Session.Exists(r.Context(), "userID") {
		a.Session.Put(r.Context(), "warning", "You must log in to see this page!")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	plans, err := a.Models.Plan.GetAll()
	if err != nil {
		a.ErrorLog.Println(err)
		return
	}

	a.render(w, r, "plans.page.gohtml", &TemplateData{
		Data: map[string]interface{}{
			"plans": plans,
		},
	})
}

func (a *App) SubscribeToPlan(w http.ResponseWriter, r *http.Request) {

}
