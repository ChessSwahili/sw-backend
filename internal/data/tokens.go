package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"backend.chesswahili.com/internal/validator"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset = "password-reset"

)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UUID      string    `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(UUID string, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UUID:   UUID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// A byte slice of length 16 is created using the make function to store random bytes that will be
	// used to generate the plaintext value of the token.
	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes) // fill slice with random bytes
	if err != nil {
		return nil, err
	}

	//encoded randomly generated byte slice using the base32.StdEncoding with no padding to create the plaintext value of the token.
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:] // Note we have use [:] since our hash func returns an array hence we convert array to slice with [:]
	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m TokenModel) New(UUID string, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(UUID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens (hash, uuid, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UUID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m TokenModel) DeleteAllForUser(scope string, UUID string) error {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND uuid = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, query, scope, UUID)
	return err
}
