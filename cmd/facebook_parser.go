package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/binsabit/dataleak/internal/data"
)

func (app *application) ParseFacebook(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("Kazakhstan.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {

		userStr := strings.Split(scanner.Text(), ":")
		fmt.Println(userStr)

		tempUser := data.SearchData{
			Phone:        userStr[0],
			FirstName:    userStr[2],
			LastName:     userStr[3],
			Gender:       userStr[4],
			Location:     fmt.Sprintf("%s %s", userStr[5], userStr[6]),
			FamilyStatus: userStr[7],
			Occupation:   userStr[8],
		}
		err = app.models.Search.Insert(tempUser)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
