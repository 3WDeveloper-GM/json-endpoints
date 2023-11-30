package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UsrID     int64     `json:"-"`
	Expity    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UsrID:  userID,
		Expity: time.Now().Add(ttl),
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

func ValidateTokenPlaintext(v *validator.Validator, tokenPlainText string) {

	v.Check(tokenPlainText != "", "token", messageMustProvide)

	var exactCharAmount = 26
	v.Check(len(tokenPlainText) == exactCharAmount, "token", fmt.Sprintf("must be exactly %v bytes long", exactCharAmount))
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(UserID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(UserID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)
	`
	args := []interface{}{token.Hash, token.UsrID, token.Expity, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteForAllUser(scope string, userID int64) error {
	query := `
		DELETE FROM tokens 
		WHERE scope = $1 and user_id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
