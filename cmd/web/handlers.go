package main

import (
	"fmt"
	"github.com/phpdave11/gofpdf"
	"github.com/phpdave11/gofpdf/contrib/gofpdi"
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"html/template"
	"net/http"
	"strconv"
	"time"
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
	planIDStr := r.URL.Query().Get("id")
	planID, _ := strconv.Atoi(planIDStr)

	plan, err := a.Models.Plan.GetOne(planID)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to find plan.")
		http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
		return
	}

	user, ok := a.Session.Get(r.Context(), "user").(models.User)
	if !ok {
		a.ErrorLog.Println("cannot get user from session")
		a.Session.Put(r.Context(), "error", "Log in first!")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	a.sendInvoice(user, plan)
	a.sendManual(user, plan)

	err = a.Models.Plan.SubscribeUserToPlan(user, *plan)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to subscribe.")
		http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
		return
	}

	u, err := a.Models.User.GetOne(user.ID)
	if err != nil {
		a.ErrorLog.Println(err)
		a.Session.Put(r.Context(), "error", "Unable to get user from database.")
		http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
		return
	}

	a.Session.Put(r.Context(), "user", u)
	a.Session.Put(r.Context(), "flash", "Subscription ordered - check your email for confirmation.")
	http.Redirect(w, r, "/members/plans", http.StatusSeeOther)
}

func (a *App) sendInvoice(user models.User, plan *models.Plan) {
	a.WaitGroup.Add(1)
	go func() {
		defer a.WaitGroup.Done()

		invoice, err := a.generateInvoice(user, plan)
		if err != nil {
			a.ErrorChan <- err
			return
		}

		a.sendEmail(Message{
			ToAddress: user.Email,
			Subject:   "Your invoice",
			Data:      invoice,
			Template:  "mail",
		})
	}()
}

func (a *App) generateInvoice(user models.User, plan *models.Plan) (string, error) {
	time.Sleep(3 * time.Second)
	return fmt.Sprintf("Invoice for %s %s: %s", user.FirstName, user.LastName, plan.PlanAmountFormatted), nil
}

func (a *App) sendManual(user models.User, plan *models.Plan) {
	a.WaitGroup.Add(1)
	go func() {
		defer a.WaitGroup.Done()

		filename := fmt.Sprintf("./tmp/%d_manual.pdf", user.ID)
		pdf := a.generateManual(user, plan)
		err := pdf.OutputFileAndClose(filename)
		if err != nil {
			a.ErrorChan <- err
			return
		}

		a.sendEmail(Message{
			ToAddress: user.Email,
			Subject:   "Your manual",
			Data:      "Your user manual is attached.",
			AttachmentsMap: map[string]string{
				"Manual.pdf": filename,
			},
			Template: "mail",
		})
	}()
}

func (a *App) generateManual(user models.User, plan *models.Plan) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(10, 13, 10)

	time.Sleep(6 * time.Second)
	importer := gofpdi.NewImporter()
	t := importer.ImportPage(pdf, "./pdf/manual.pdf", 1, "/MediaBox")
	pdf.AddPage()

	importer.UseImportedTemplate(pdf, t, 0, 0, 215.9, 0)
	pdf.SetX(75)
	pdf.SetY(150)

	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(0, 4, fmt.Sprintf("%s %s", user.FirstName, user.LastName), "", "C", false)
	pdf.Ln(5)

	pdf.MultiCell(0, 4, fmt.Sprintf("%s User Guide", plan.PlanName), "", "C", false)
	return pdf
}
