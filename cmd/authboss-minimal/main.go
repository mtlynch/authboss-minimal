package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/volatiletech/authboss/v3"
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

	var myDatabaseImplementation authboss.ServerStorer
	var mySessionImplementation authboss.ClientStateReadWriter
	var myCookieImplementation authboss.ClientStateReadWriter

	myDatabaseImplementation = inMemoryStore{}

	ab.Config.Storage.Server = myDatabaseImplementation
	ab.Config.Storage.SessionState = mySessionImplementation
	ab.Config.Storage.CookieState = myCookieImplementation

	ab.Config.Paths.Mount = "/authboss"
	ab.Config.Paths.RootURL = fmt.Sprintf("http://0.0.0.0:%s", port)

	readJson := false
	useUsername := false
	defaults.SetCore(&ab.Config, readJson, useUsername)

	if err := ab.Init(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Use(ab.LoadClientStateMiddleware)

	views := router.PathPrefix("/").Subrouter()
	views.HandleFunc("/", publicHello)
	views.Handle(ab.Config.Paths.Mount, http.StripPrefix(ab.Config.Paths.Mount, ab.Config.Core.Router))

	authenticatedViews := router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(authboss.Middleware2(ab, authboss.RequireNone, authboss.RespondUnauthorized))
	authenticatedViews.HandleFunc("/private", privateHello)

	log.Printf("Listening on %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
