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
	ID       entity.ID `json:"id"` // ID é criado por um pacote na pasta pkg/entity
	Name     string    `json:"name"`
	Email    string    `json:"email" gorm:"uniqueIndex;size:255"`
	Password string    `json:"-"` // guarda o hash bcrypt; nunca serializar
}

/*
NewUser cria uma nova instância de User aplicando normalização básica de entrada,
validações mínimas de segurança e persistindo apenas o hash da senha.

Fluxo:
 1. Normaliza os parâmetros:
    - name: TrimSpace
    - email: TrimSpace + strings.ToLower
    - plainPassword: TrimSpace
 2. Valida o e-mail via isValidEmail; se inválido, retorna ErrInvalidEmail.
 3. Valida a senha de forma mínima (apenas não vazia após trim); se vazia, retorna ErrWeakPassword.
 4. Gera o hash da senha via hashPassword; se falhar, propaga o erro retornado pela função.
 5. Constrói e retorna um User com:
    - ID gerado via entity.NewId()
    - Name e Email normalizados
    - Password contendo o hash (nunca a senha em texto puro)

Parâmetros:
  - name: nome do usuário (normalizado com trim).
  - email: e-mail do usuário (normalizado com trim e lower-case).
  - plainPassword: senha em texto puro (somente para geração do hash; não é armazenada).

Retorno:
  - (*User, nil) em caso de sucesso.
  - (nil, ErrInvalidEmail) se o e-mail for inválido.
  - (nil, ErrWeakPassword) se a senha for vazia após normalização.
  - (nil, err) se ocorrer erro ao gerar o hash da senha.
*/
func NewUser(name, email, plainPassword string) (*User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))
	plainPassword = strings.TrimSpace(plainPassword)

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if plainPassword == "" { // No momento só verifica se não é vazio
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

/*
SetPassword atualiza o campo Password do usuário a partir de uma senha em texto puro,
aplicando normalização básica e armazenando apenas o hash resultante.

Fluxo:
 1. Normaliza plainPassword com strings.TrimSpace.
 2. Valida a senha de forma mínima (apenas não vazia após trim); se vazia, retorna ErrWeakPassword.
 3. Gera o hash da senha via hashPassword; se falhar, propaga o erro retornado pela função.
 4. Atualiza u.Password com o hash gerado.

Parâmetros:
  - plainPassword: senha em texto puro (usada somente para gerar o hash; não é armazenada).

Retorno:
  - nil em caso de sucesso.
  - ErrWeakPassword se a senha for vazia após normalização.
  - err se ocorrer erro ao gerar o hash da senha.
*/
func (u *User) SetPassword(plainPassword string) error {
	plainPassword = strings.TrimSpace(plainPassword)
	if plainPassword == "" { // No momento só verifica se não é vazio
		return ErrWeakPassword
	}
	hash, err := hashPassword(plainPassword)
	if err != nil {
		return err
	}
	u.Password = hash
	return nil
}

/*
CheckPassword verifica se a senha em texto puro informada corresponde ao hash atualmente
armazenado em u.Password.

Fluxo:
 1. Faz checagens defensivas:
    - se u == nil, retorna false.
    - se u.Password estiver vazio, retorna false.
 2. Compara a senha em texto puro com o hash usando bcrypt.CompareHashAndPassword.
    - Retorna true se a comparação não retornar erro.
    - Retorna false caso contrário.

Parâmetros:
  - plainPassword: senha em texto puro a ser validada contra o hash armazenado.

Retorno:
  - true se plainPassword corresponder ao hash em u.Password.
  - false se o usuário for nil, se não houver hash armazenado, ou se a comparação falhar.
*/
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
