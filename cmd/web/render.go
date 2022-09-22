package main

import (
	"fmt"
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"html/template"
	"net/http"
	"time"
)

var (
	pathToTemplates = "./cmd/web/templates"
)

type TemplateData struct {
	StringMap     map[string]string
	FloatMap      map[string]float64
	IntMap        map[string]int
	Data          map[string]interface{}
	Flash         string
	Warning       string
	Error         string
	Now           time.Time
	Authenticated bool
	User          *models.User
}

func (a *App) render(w http.ResponseWriter, r *http.Request, name string, data *TemplateData) {
	partials := []string{
		fmt.Sprintf("%s/base.layout.gohtml", pathToTemplates),
		fmt.Sprintf("%s/header.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/navbar.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/footer.partial.gohtml", pathToTemplates),
		fmt.Sprintf("%s/alerts.partial.gohtml", pathToTemplates),
	}

	var templateSlice []string
	templateSlice = append(templateSlice, fmt.Sprintf("%s/%s", pathToTemplates, name))

	for _, x := range partials {
		templateSlice = append(templateSlice, x)
	}

	if data == nil {
		data = &TemplateData{}
	}

	parsedTemplate, err := template.ParseFiles(templateSlice...)
	if err != nil {
		a.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = parsedTemplate.Execute(w, a.AddDefaultData(data, r))
	if err != nil {
		a.ErrorLog.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) AddDefaultData(data *TemplateData, r *http.Request) *TemplateData {
	data.Flash = a.Session.PopString(r.Context(), "flash")
	data.Warning = a.Session.PopString(r.Context(), "warning")
	data.Error = a.Session.PopString(r.Context(), "error")
	data.Authenticated = a.IsAuthenticated(r)

	if data.Authenticated {
		user, ok := a.Session.Get(r.Context(), "user").(models.User)
		if !ok {
			a.ErrorLog.Println("cannot get user from session")
		} else {
			data.User = &user
		}
	}

	data.Now = time.Now()
	return data
}

func (a *App) IsAuthenticated(r *http.Request) bool {
	return a.Session.Exists(r.Context(), "userID")
}
