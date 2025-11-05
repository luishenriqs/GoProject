// internal/entity/user.go
package entity

import (
	"errors"
	"net/mail"
	"strings"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRequired     = errors.New("required value missing")
	ErrInvalidEmail = errors.New("invalid email")
	ErrWeakPassword = errors.New("password must not be blank")
)

// User representa o agregado de usuário.
// Campo Password armazena o hash bcrypt e é omitido do JSON.
type User struct {
	ID       entity.ID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email" gorm:"uniqueIndex"` // índice único
	Password string    `json:"-"`                        // guarda o hash bcrypt; nunca serializar
}

// NewUser cria um usuário novo, validando campos mínimos e
// armazenando o hash da senha em Password.
func NewUser(name, email, plainPassword string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if strings.TrimSpace(plainPassword) == "" {
		return nil, ErrWeakPassword
	}

	hash, err := hashPassword(plainPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       entity.NewId(),
		Name:     name,
		Email:    email,
		Password: hash,
	}, nil
}

func (u *User) SetPassword(plainPassword string) error {
	if strings.TrimSpace(plainPassword) == "" {
		return ErrWeakPassword
	}
	hash, err := hashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

func (u *User) CheckPassword(plainPassword string) bool {
	if u == nil || u.Password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword)) == nil
}

func isValidEmail(e string) bool {
	parsed, err := mail.ParseAddress(e)
	return err == nil && parsed.Address != ""
}

func hashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
