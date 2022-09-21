package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"testing"
)

var expectedRoutes = []string{
	"/",
	"/login",
	"/login",
	"/logout",
	"/register",
	"/register",
	"/activate",
	"/members/plans",
	"/members/subscribe",
}

func Test_RoutesExist(t *testing.T) {
	testRoutes := testApp.routes()
	chiRoutes := testRoutes.(chi.Router)

	for _, route := range expectedRoutes {
		routeExists(t, chiRoutes, route)
	}
}

func routeExists(t *testing.T, routes chi.Router, expectedRoute string) {
	found := false

	_ = chi.Walk(routes, func(method string,
		chiRoute string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler) error {
		if expectedRoute == chiRoute {
			found = true
		}
		return nil
	})

	if !found {
		t.Errorf("Route %q has not been found in registered routes.", expectedRoute)
	}
}
