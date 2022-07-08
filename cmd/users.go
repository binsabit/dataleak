package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/binsabit/dataleak/internal/data"
	"github.com/binsabit/dataleak/internal/validator"
)

func (app *application) Parse(w http.ResponseWriter, r *http.Request) {

}

func (app *application) SignIn(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	matched, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	if !matched {
		fmt.Print("Not matched")
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication": token}, nil)
	if err != nil {

		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) SignUp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Processing Post request")
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	user := &data.User{
		Email: input.Email,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address alreadt exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) Logout(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)
	fmt.Println("logout")
	if user.IsAnonymous() {
		app.authenticationRequiredResponse(w, r)
		return
	}
	fmt.Println("passed athentication")
	fmt.Println(user)
	err := app.models.Tokens.DeleteAllForUser("authentication", user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"message": "user logged out"}, nil)

}

func (app *application) SearchForData(w http.ResponseWriter, r *http.Request) {
	user := app.contextGetUser(r)

	if user.IsAnonymous() {
		app.authenticationRequiredResponse(w, r)
		return
	}
	// fmt.Println("Here")
	var input struct {
		PlainText string `json:"search_item"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	e := validator.New()
	p := validator.New()
	fmt.Println(input.PlainText)

	var d []data.FacebookParser
	passed := 0
	if data.ValidateEmail(e, input.PlainText); e.Valid() {

		passed++
	}
	if data.ValidatePhone(p, input.PlainText); p.Valid() {
		passed++
	}
	d, err = app.models.Search.GetInfoOf(input.PlainText)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	app.writeJSON(w, http.StatusOK, envelope{"data": envelope{"arr": d}}, nil)
}
