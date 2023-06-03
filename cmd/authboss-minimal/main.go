package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	"github.com/volatiletech/authboss/v3/defaults"
)

func publicHello(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprint(w, "Hello, world!"); err != nil {
		panic(err)
	}
}

func privateHello(w http.ResponseWriter, r *http.Request) {
	if _, err := fmt.Fprint(w, "Hello, only authenticated users!"); err != nil {
		panic(err)
	}
}

type inMemoryStore struct {
	Users map[string]authboss.User
}

// Load will look up the user based on the passed the PrimaryID.
func (s inMemoryStore) Load(_ context.Context, key string) (authboss.User, error) {
	log.Printf("loading user with key %v", key)
	u, ok := s.Users[key]
	if !ok {
		return nil, authboss.ErrUserNotFound
	}

	return u, nil
}

// Save persists the user in the database.
func (s inMemoryStore) Save(_ context.Context, user authboss.User) error {
	key := user.GetPID()

	log.Printf("saving user with key %v", key)

	s.Users[key] = user

	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	ab := authboss.New()

	readJson := true
	useUsername := false
	defaults.SetCore(&ab.Config, readJson, useUsername)

	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}

	cookieStoreKey, err := base64.StdEncoding.DecodeString(`NpEPi8pEjKVjLGJ6kYCS+VTCzi6BUuDzU0wrwXyf5uDPArtlofn2AG6aTMiPmN3C909rsEWMNqJqhIVPGP3Exg==`)
	if err != nil {
		panic(err)
	}

	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.HTTPOnly = false
	cookieStore.Secure = false
	sessionStoreKey, err := base64.StdEncoding.DecodeString(`AbfYwmmt8UCwUuhd9qvfNA9UCuN1cVcKJN1ofbiky6xCyyBj20whe40rJa3Su0WOWLWcPpO1taqJdsEI/65+JA==`)
	if err != nil {
		panic(err)
	}
	const sessionCookieName = "ab_blog"
	sessionStore := abclientstate.NewSessionStorer(sessionCookieName, sessionStoreKey, nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false
	cstore.MaxAge(int((30 * 24 * time.Hour) / time.Second))

	ab.Config.Storage.Server = inMemoryStore{}
	ab.Config.Storage.SessionState = sessionStore
	ab.Config.Storage.CookieState = cookieStore

	ab.Config.Paths.Mount = "/auth"
	ab.Config.Paths.RootURL = fmt.Sprintf("http://0.0.0.0:%s", port)

	if err := ab.Init(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Use(ab.LoadClientStateMiddleware)

	router.PathPrefix(ab.Config.Paths.Mount + "/login").Handler(http.StripPrefix(ab.Config.Paths.Mount, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("we're in the authboss subrouter!")
		log.Printf("request=%s", r.URL.Path)
		ab.Config.Core.Router.ServeHTTP(w, r)
	}))).Methods(http.MethodPost)

	views := router.PathPrefix("/").Subrouter()
	views.HandleFunc("/", publicHello)

	authenticatedViews := router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(authboss.Middleware2(ab, authboss.RequireNone, authboss.RespondUnauthorized))
	authenticatedViews.HandleFunc("/private", privateHello)

	log.Printf("Listening on %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
