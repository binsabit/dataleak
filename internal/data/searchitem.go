package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/binsabit/dataleak/internal/validator"
)

type SearchItem struct {
	Plaintext string `json:"plaintext"`
	Type      string `json:"type"`
}

type FacebookParser struct {
	ID           string `json:"record_id"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FirstName    string `json:"firstname"`
	LastName     string `json:"lastname"`
	Gender       string `json:"gender"`
	Location     string `json:"location"`
	FamilyStatus string `json:"family_status"`
	Occupation   string `json:"occupation"`
}

func ValidatePhone(v *validator.Validator, phone string) {
	v.Check(phone != "", "phone", "must be provided")
	v.Check(validator.Matches(phone, validator.PhoneRX), "phone", "must be phone number")
}

type SearchItemModel struct {
	DB *sql.DB
}

func (m SearchItemModel) GetInfoOf(s string) ([]FacebookParser, error) {
	query := `SELECT * FROM dataleak WHERE email=$1 OR phone=$1`
	rows, err := m.DB.Query(query, s)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var data []FacebookParser
	for rows.Next() {
		var temp FacebookParser
		err = rows.Scan(
			&temp.ID,
			&temp.Phone,
			&temp.FirstName,
			&temp.LastName,
			&temp.Gender,
			&temp.Location,
			&temp.FamilyStatus,
			&temp.Occupation,
			&temp.Email)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, ErrRecordNotFound
			default:
				return nil, err
			}
		}
		data = append(data, temp)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (m SearchItemModel) Insert(f FacebookParser) error {
	query := `INSERT INTO dataleak (phone, firstname,lastname,gender,location,familystatus,occupation,email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id`
	args := []interface{}{f.Phone, f.FirstName, f.LastName, f.Gender, f.Location, f.FamilyStatus, f.Occupation, f.Email}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&f.ID)
	if err != nil {
		switch {
		default:
			return err
		}
	}

	return nil
}
