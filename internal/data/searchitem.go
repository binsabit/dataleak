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

type SearchData struct {
	ID           string `json:"record_id"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FirstName    string `json:"firstname"`
	LastName     string `json:"lastname"`
	Gender       string `json:"gender"`
	Location     string `json:"location"`
	FamilyStatus string `json:"family_status"`
	Occupation   string `json:"occupation"`
	Source       string `json:"source"`
}

func (s SearchData) IsEmpty() bool {
	return s == SearchData{}
}
func ValidatePhone(v *validator.Validator, phone string) {
	v.Check(phone != "", "phone", "must be provided")
	v.Check(validator.Matches(phone, validator.PhoneRX), "phone", "must be phone number")
}

type SearchItemModel struct {
	DB *sql.DB
}

func (m SearchItemModel) GetInfoByEmail(email string) ([]SearchData, error) {
	query := `SELECT * FROM dataleak WHERE email=$1`
	rows, err := m.DB.Query(query, email)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var data []SearchData
	for rows.Next() {
		var temp SearchData
		err = rows.Scan(
			&temp.ID,
			&temp.Email,
			&temp.Phone,
			&temp.FirstName,
			&temp.LastName,
			&temp.Gender,
			&temp.Location,
			&temp.FamilyStatus,
			&temp.Occupation,
			&temp.Source,
		)
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

func (m SearchItemModel) GetInfoOf(s string) ([]SearchData, error) {
	query := `SELECT * FROM dataleak WHERE email=$1 OR phone LIKE '%' || $1 || '%'`
	rows, err := m.DB.Query(query, s)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var data []SearchData
	for rows.Next() {
		var temp SearchData
		err = rows.Scan(
			&temp.ID,
			&temp.Email,
			&temp.Phone,
			&temp.FirstName,
			&temp.LastName,
			&temp.Gender,
			&temp.Location,
			&temp.FamilyStatus,
			&temp.Occupation,
			&temp.Source,
		)
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

func (m SearchItemModel) Insert(f SearchData) error {
	query := `INSERT INTO dataleak (phone, firstname,lastname,gender,location,family_status,occupation,email,source)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`
	args := []interface{}{f.Phone, f.FirstName, f.LastName, f.Gender, f.Location, f.FamilyStatus, f.Occupation, f.Email, f.Source}
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
