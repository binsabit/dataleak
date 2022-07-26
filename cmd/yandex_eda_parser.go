package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/binsabit/dataleak/internal/data"
)

func (app *application) ParseYandexEda(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open("yandex_eda.csv")
	if err != nil {
		log.Fatal(err)
	}
	wr, err := os.Create("data.txt")
	if err != nil {
		log.Fatal("cannot create log file")
	}

	defer wr.Close()

	o := 0
	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	csvReader.Comma = '+'
	csvReader.LazyQuotes = true
	_, _ = csvReader.Read()
	// var wg sync.WaitGroup
	for {
		// go func() {

		// }()
		// defer wg.Done()
		rawRecord, err := csvReader.Read()
		// fmt.Println(rawRecord[1])
		if err == io.EOF {
			break
		}

		for _, val := range rawRecord {
			if len(val) < 1 {
				continue
			}
			o++
			s := parseRecord(val)
			if s.IsEmpty() {
				continue
			}
			_, err = wr.WriteString(fmt.Sprintf("%s;%s;%s;%s;%s\n", s.Phone, s.Email, s.FirstName, s.LastName, s.Location))
			if err != nil {
				continue
			}
			err = app.models.Search.Insert(s)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%v + o=%d\n", val, o)

		}

	}

	fmt.Println(o)
}

func parseRecord(record string) data.SearchData {
	s := strings.Split(record, ";")
	// s1 := strings.Split(record, ",")
	temp := data.SearchData{}

	if len(s) < 5 {
		fmt.Println("Break", record)
		return temp
	}
	temp.Phone = s[0]
	temp.FirstName = s[1]
	temp.LastName = s[2]
	temp.Email = s[3]
	temp.Location = s[4]
	temp.Source = "Yandex eda 2022 leak"
	fmt.Println(s, len(s))
	return temp
}
