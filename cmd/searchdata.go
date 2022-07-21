package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/binsabit/dataleak/internal/data"
	"github.com/binsabit/dataleak/internal/validator"
)

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

	var d []data.SearchData
	passed := 0
	if data.ValidateEmail(e, input.PlainText); e.Valid() {

		passed++
	}
	if data.ValidatePhone(p, input.PlainText); p.Valid() {
		passed++
	}
	if passed == 0 {
		app.badRequestResponse(w, r, fmt.Errorf("Bad data entry types are not valid"))
	}
	input.PlainText = input.PlainText[1:]
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
	app.writeJSON(w, http.StatusOK, envelope{"data": d}, nil)
}
