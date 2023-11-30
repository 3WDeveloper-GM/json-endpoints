package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/3WDeveloper-GM/json-endpoints/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var AnonUser = &User{}

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	Passwd *string
	Hash   []byte
}

func (u *User) IsAnonymous() bool {
	return u == AnonUser
}

func (p *password) Set(Passwd string) error {
	Hash, err := bcrypt.GenerateFromPassword([]byte(Passwd), 12)
	if err != nil {
		return err
	}

	p.Passwd = &Passwd
	p.Hash = Hash

	return nil
}

func (p *password) Matches(Passwd string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(Passwd))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

var messageMustProvide = "must provide a value, entry cannot be empty"
var messageMBAL = "must be at least"
var messageMNBMT = "must not be more than"

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", messageMustProvide)

	var message = "must be a valid email address"
	v.Check(validator.Matches(email, validator.EMailMustCompile), "email", message)
}

func ValidatePasswordPlaintext(v *validator.Validator, plaintext string) {
	v.Check(plaintext != "", "password", messageMustProvide)

	var minimumPasschar = 8
	var maximumPasschar = 72

	v.Check(len(plaintext) >= minimumPasschar, "password", fmt.Sprintf(messageMBAL+" %v bytes long.", minimumPasschar))
	v.Check(len(plaintext) <= maximumPasschar, "password", fmt.Sprintf(messageMNBMT+" %v bytes long.", maximumPasschar))
}

func ValidateUser(v *validator.Validator, usr *User) {
	v.Check(usr.Name != "", "username", messageMustProvide)

	var maximumUsrNameChar = 500
	v.Check(len(usr.Name) <= maximumUsrNameChar, "username", fmt.Sprintf(messageMNBMT+" %v characters long", messageMNBMT))

	ValidateEmail(v, usr.Email)

	if usr.Password.Passwd != nil {
		ValidatePasswordPlaintext(v, *usr.Password.Passwd)
	}

	if usr.Password.Hash == nil {
		panic("missing password Hash for user")
	}
}

var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_Hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, email, password_Hash, activated, version
		FROM users
		WHERE email = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		SET name = $1, email = $2, password_Hash = $3, activated = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version
	`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.Hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.Version,
	)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint key "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		FROM users
		INNER JOIN tokens
		ON users.id = tokens.user_id
		WHERE tokens.hash = $1
		AND tokens.scope = $2
		AND tokens.expiry > $3
	`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
