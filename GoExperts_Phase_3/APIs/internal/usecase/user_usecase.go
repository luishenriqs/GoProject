package usecase

import (
	"errors"
	"strings"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/infra/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// ErrInvalidCredentials representa credenciais inválidas (email/senha incorretos).
var ErrInvalidCredentials = errors.New("invalid credentials")

// UserUseCase concentra regras de aplicação/orquestração de User.
// Aqui não existe lógica HTTP, apenas coordenação: domínio + repositório.
type UserUseCase struct {
	UserRepo database.UserInterface
}

// NewUserUseCase cria o usecase de User injetando a interface do repositório.
func NewUserUseCase(repo database.UserInterface) *UserUseCase {
	return &UserUseCase{UserRepo: repo}
}

/*
RegisterUser registra um novo usuário a partir de um payload de criação,
delegando a validação e a geração do hash de senha para a camada de entidade,
e persistindo o usuário via repositório.

Fluxo:
 1. Cria a entidade de usuário chamando entity.NewUser(input.Name, input.Email, input.Password).
    - Se a criação/validação falhar, retorna (nil, err).
 2. Persiste o usuário no repositório via uc.UserRepo.Create(u).
    - Se a persistência falhar, retorna (nil, err).
 3. Retorna a entidade criada em caso de sucesso.

Parâmetros:
  - input: dto.CreateUserInput contendo Name, Email e Password (senha em texto puro usada apenas
    para criação do hash dentro da entidade).

Retorno:
  - (*entity.User, nil) em caso de sucesso.
  - (nil, err) se a criação/validação do usuário falhar ou se a persistência no repositório falhar.
*/
func (uc *UserUseCase) RegisterUser(input dto.CreateUserInput) (*entity.User, error) {
	u, err := entity.NewUser(input.Name, input.Email, input.Password)
	if err != nil {
		return nil, err
	}

	if err := uc.UserRepo.Create(u); err != nil {
		return nil, err
	}

	return u, nil
}

/*
Login autentica um usuário a partir de e-mail e senha, retornando o ID do usuário
quando as credenciais forem válidas.

Fluxo:
 1. Normaliza o e-mail de entrada:
    - strings.TrimSpace
    - strings.ToLower
 2. Busca o usuário no repositório por e-mail via uc.UserRepo.FindByEmail(normalizedEmail).
    - Se o erro for gorm.ErrRecordNotFound, retorna (pkgentity.ID{}, gorm.ErrRecordNotFound).
    - Se houver outro erro, retorna (pkgentity.ID{}, err).
 3. Valida a senha usando u.CheckPassword(password).
    - Se a senha não corresponder ao hash armazenado, retorna (pkgentity.ID{}, ErrInvalidCredentials).
 4. Retorna o u.ID em caso de sucesso.

Parâmetros:
  - email: e-mail do usuário (normalizado internamente).
  - password: senha em texto puro a ser validada contra o hash armazenado.

Retorno:
  - (pkgentity.ID, nil) se a autenticação for bem-sucedida.
  - (pkgentity.ID{}, gorm.ErrRecordNotFound) se não existir usuário com o e-mail informado.
  - (pkgentity.ID{}, ErrInvalidCredentials) se o usuário existir, mas a senha estiver incorreta.
  - (pkgentity.ID{}, err) para erros inesperados ao buscar o usuário no repositório.
*/
func (uc *UserUseCase) Login(email, password string) (pkgentity.ID, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	u, err := uc.UserRepo.FindByEmail(normalizedEmail)
	if err != nil {
		// Propaga not-found do GORM para o handler decidir o status code,
		// mas a autenticação em geral deve responder "credenciais inválidas".
		// Nesta etapa, mantemos simples e retornamos ErrInvalidCredentials
		// apenas quando o usuário existe mas a senha é incorreta.
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pkgentity.ID{}, gorm.ErrRecordNotFound
		}
		return pkgentity.ID{}, err
	}

	if !u.CheckPassword(password) {
		return pkgentity.ID{}, ErrInvalidCredentials
	}

	return u.ID, nil
}

// GetMe carrega o usuário autenticado pelo ID.
func (uc *UserUseCase) GetMe(userID pkgentity.ID) (*entity.User, error) {
	return uc.UserRepo.FindByID(userID)
}

/*
UpdateMe atualiza os dados do próprio usuário (identificado por userID),
aplicando alterações parciais (patch) em Name e/ou Password e persistindo
o resultado no repositório.

Fluxo:
 1. Busca o usuário no repositório por ID via uc.UserRepo.FindByID(userID).
    - Se falhar, retorna (nil, err).
 2. Se input não contém nenhum campo (Name == nil e Password == nil), não altera nada (no-op)
    e retorna o usuário carregado.
 3. Se Name foi fornecido:
    - Normaliza com strings.TrimSpace.
    - Se ficar vazio, retorna entity.ErrRequired.
    - Caso contrário, atualiza u.Name.
 4. Se Password foi fornecido:
    - Atualiza via u.SetPassword, que valida minimamente e armazena o hash.
    - Se falhar, retorna (nil, err).
 5. Persiste as alterações via uc.UserRepo.Update(u).
    - Se falhar, retorna (nil, err).
 6. Retorna o usuário atualizado.

Parâmetros:
  - userID: ID do usuário a ser atualizado.
  - input: dto.UpdateUserInput com campos opcionais (ponteiros) para atualização parcial:
  - Name: *string (opcional)
  - Password: *string (opcional)

Retorno:
  - (*entity.User, nil) em caso de sucesso (incluindo no-op quando nada é enviado).
  - (nil, entity.ErrRequired) se Name for enviado e resultar em string vazia após trim.
  - (nil, err) para falhas de busca por ID, falhas ao setar senha ou falhas ao persistir no repositório.
*/
func (uc *UserUseCase) UpdateMe(userID pkgentity.ID, input dto.UpdateUserInput) (*entity.User, error) {
	u, err := uc.UserRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Se nada foi enviado, não altera nada (no-op).
	if input.Name == nil && input.Password == nil {
		return u, nil
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, entity.ErrRequired
		}
		u.Name = name
	}

	if input.Password != nil {
		if err := u.SetPassword(*input.Password); err != nil {
			return nil, err
		}
	}

	if err := uc.UserRepo.Update(u); err != nil {
		return nil, err
	}

	return u, nil
}

// DeleteMe remove o usuário autenticado.
func (uc *UserUseCase) DeleteMe(userID pkgentity.ID) error {
	return uc.UserRepo.Delete(userID)
}
