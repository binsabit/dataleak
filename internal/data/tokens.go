package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/binsabit/dataleak/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"."`
	UserID    int64     `json:"."`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"scope"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil

}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must ne 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (t TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {

	query := `
		INSERT INTO tokens (hash, user_id, expiry, scope)
		VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}
func (m TokenModel) GetAllForUser(user *User) ([]*Token, error) {
	query := `SELECT hash, user_id, expiry, scope
			FROM tokens
			WHERE user_id = $1`
	rows, err := m.DB.Query(query, user.ID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var tokens []*Token
	for rows.Next() {
		var tempToken Token
		err = rows.Scan(&tempToken.Hash, &tempToken.UserID, &tempToken.Expiry, &tempToken.Scope)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, &tempToken)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return tokens, nil
}
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens
		WHERE scope = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
