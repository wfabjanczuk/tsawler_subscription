package main

import (
	"github.com/wfabjanczuk/tsawler_subscription/cmd/web/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var pageTests = []struct {
	name               string
	url                string
	handler            http.HandlerFunc
	sessionData        map[string]interface{}
	expectedStatusCode int
	expectedHtml       string
}{
	{
		name:               "home",
		url:                "/",
		handler:            testApp.HomePage,
		expectedStatusCode: http.StatusOK,
	},
	{
		name:               "login",
		url:                "/login",
		handler:            testApp.LoginPage,
		expectedStatusCode: http.StatusOK,
		expectedHtml:       `<h1 class="mt-5">Login</h1>`,
	},
	{
		name:    "logout",
		url:     "/logout",
		handler: testApp.Logout,
		sessionData: map[string]interface{}{
			"userID": 1,
			"user":   models.User{},
		},
		expectedStatusCode: http.StatusSeeOther,
	},
}

func TestApp_Pages(t *testing.T) {
	for _, pageTest := range pageTests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", pageTest.url, nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		pageTest.handler.ServeHTTP(rr, req)

		if len(pageTest.sessionData) > 0 {
			for key, value := range pageTest.sessionData {
				testApp.Session.Put(ctx, key, value)
			}
		}

		if rr.Code != pageTest.expectedStatusCode {
			t.Errorf("Page test %q failed: expected %d but got %d", pageTest.name, pageTest.expectedStatusCode, rr.Code)
		}

		if len(pageTest.expectedHtml) > 0 {
			html := rr.Body.String()

			if !strings.Contains(html, pageTest.expectedHtml) {
				t.Errorf("Page test %q failed: unable to find %s in response body", pageTest.name, pageTest.expectedHtml)
			}
		}
	}
}

func TestApp_PostLoginPage(t *testing.T) {
	formData := url.Values{
		"email":    {"admin@example.com"},
		"password": {"password"},
	}
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(formData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	handler := http.HandlerFunc(testApp.PostLoginPage)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Page test %q failed: expected %d but got %d", "PostLogin", http.StatusSeeOther, rr.Code)
	}

	if !testApp.Session.Exists(ctx, "userID") {
		t.Errorf("Page test %q failed: unable to find userID in session", "PostLogin")
	}
}

func TestApp_SubscribeToPlan(t *testing.T) {
	rr := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/members/subscribe?id=1", nil)
	ctx := getCtx(req)
	req = req.WithContext(ctx)

	testApp.Session.Put(ctx, "userID", 1)
	testApp.Session.Put(ctx, "user", models.User{
		ID:        1,
		Email:     "admin@example.com",
		FirstName: "Admin",
		LastName:  "User",
		Active:    1,
	})

	handler := http.HandlerFunc(testApp.SubscribeToPlan)
	handler.ServeHTTP(rr, req)

	testApp.WaitGroup.Wait()

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Page test %q failed: expected %d but got %d", "SubscribeToPlan", http.StatusSeeOther, rr.Code)
	}
}
