package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApp_AddDefaultData(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "flash", "flash")
	testApp.Session.Put(ctx, "warning", "warning")
	testApp.Session.Put(ctx, "error", "error")

	templateData := testApp.AddDefaultData(&TemplateData{}, req)

	if templateData.Flash != "flash" {
		t.Error("failed to add flash data")
	}

	if templateData.Warning != "warning" {
		t.Error("failed to add warning data")
	}

	if templateData.Error != "error" {
		t.Error("failed to add error data")
	}
}

func TestApp_IsAuthenticated(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	auth := testApp.IsAuthenticated(req)
	if auth {
		t.Error("unauthenticated request returns true for IsAuthenticated")
	}

	testApp.Session.Put(ctx, "userID", 1)
	auth = testApp.IsAuthenticated(req)
	if !auth {
		t.Error("authenticated request returns false for IsAuthenticated")
	}
}

func TestConfig_render(t *testing.T) {
	pathToTemplates = "./templates"

	rr := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.render(rr, req, "home.page.gohtml", &TemplateData{})

	if rr.Code != http.StatusOK {
		t.Error("failed to render page")
	}
}
