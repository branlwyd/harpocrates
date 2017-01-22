// Package handler provides the HTTP handlers used by harpocrates.
package handler

import (
	"fmt"
	"net/http"

	"../session"
)

const (
	sessionCookieName = "harp-sid"
)

func NewContent(sh *session.Handler) (http.Handler, error) {
	mux := http.NewServeMux()

	// Static content handlers.
	mux.Handle("/", newFiltered("/", http.RedirectHandler("/p/", http.StatusFound)))

	styleHandler, err := newAsset("etc/style.css", "text/css; charset=utf-8")
	if err != nil {
		return nil, fmt.Errorf("could not create style handler: %v", err)
	}
	mux.Handle("/style.css", styleHandler)

	robotsHandler, err := newAsset("etc/robots.txt", "text/plain; charset=utf-8")
	if err != nil {
		return nil, fmt.Errorf("could not create robots handler: %v", err)
	}
	mux.Handle("/robots.txt", robotsHandler)

	u2fAPIHandler, err := newAsset("etc/u2f-api.js", "application/javascript")
	if err != nil {
		return nil, fmt.Errorf("could not create U2F API handler: %v", err)
	}
	mux.Handle("/u2f-api.js", u2fAPIHandler)

	// Dynamic content handlers.
	sp, err := newSessionProvider(sh)
	if err != nil {
		return nil, fmt.Errorf("could not create session provider: %v", err)
	}

	rh, err := newRegister(sp)
	if err != nil {
		return nil, fmt.Errorf("could not create registration handler: %v", err)
	}
	mux.Handle("/register", rh)

	dh, err := newDynamic(sp)
	if err != nil {
		return nil, fmt.Errorf("could not create dynamic content handler: %v", err)
	}
	mux.Handle("/p/", dh)

	return mux, nil
}