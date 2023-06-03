package main

import (
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	ab := authboss.New()

	// Leave these as nil for now
	var myDatabaseImplementation authboss.ServerStorer
	var mySessionImplementation authboss.ClientStateReadWriter
	var myCookieImplementation authboss.ClientStateReadWriter
	ab.Config.Storage.Server = myDatabaseImplementation
	ab.Config.Storage.SessionState = mySessionImplementation
	ab.Config.Storage.CookieState = myCookieImplementation

	ab.Config.Paths.RootURL = fmt.Sprintf("http://0.0.0.0:%s", port)

	readJson := false
	useUsername := false
	defaults.SetCore(&ab.Config, readJson, useUsername)

	if err := ab.Init(); err != nil {
		panic(err)
	}

	router := mux.NewRouter()

	views := router.PathPrefix("/").Subrouter()
	views.HandleFunc("/", publicHello)

	authenticatedViews := router.PathPrefix("/").Subrouter()
	authenticatedViews.Use(authboss.Middleware2(ab, authboss.RequireNone, authboss.RespondUnauthorized))
	authenticatedViews.HandleFunc("/private", privateHello)

	log.Printf("Listening on %s", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
