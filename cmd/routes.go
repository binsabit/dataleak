package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/signup", app.SignUp).Methods("POST")
	router.HandleFunc("/api/signin", app.SignIn).Methods("POST")
	router.HandleFunc("/api/logout", app.authenticate(app.Logout)).Methods("GET")
	router.HandleFunc("/api/data", app.authenticate(app.SearchForData)).Methods("POST")
	router.HandleFunc("/api/parse", app.ParseFacebook).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(app.notFoundResponse)
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	return router
}
