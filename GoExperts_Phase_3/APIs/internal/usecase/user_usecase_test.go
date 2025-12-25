package usecase

import (
	"errors"
	"testing"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// fakeUserRepo é um repositório em memória para testar o usecase sem GORM/DB.
type fakeUserRepo struct {
	usersByID    map[pkgentity.ID]*entity.User
	usersByEmail map[string]pkgentity.ID

	createErr error
	updateErr error
	deleteErr error
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		usersByID:    make(map[pkgentity.ID]*entity.User),
		usersByEmail: make(map[string]pkgentity.ID),
	}
}

func (r *fakeUserRepo) Create(user *entity.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	// Armazena o ponteiro recebido (suficiente para testes).
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user.ID
	return nil
}

func (r *fakeUserRepo) FindByEmail(email string) (*entity.User, error) {
	id, ok := r.usersByEmail[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	u := r.usersByID[id]
	if u == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) FindByID(id pkgentity.ID) (*entity.User, error) {
	u, ok := r.usersByID[id]
	if !ok || u == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) Update(user *entity.User) error {
	if r.updateErr != nil {
		return r.updateErr
	}
	_, ok := r.usersByID[user.ID]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user.ID
	return nil
}

func (r *fakeUserRepo) Delete(id pkgentity.ID) error {
	if r.deleteErr != nil {
		return r.deleteErr
	}
	u, ok := r.usersByID[id]
	if !ok || u == nil {
		return gorm.ErrRecordNotFound
	}
	delete(r.usersByEmail, u.Email)
	delete(r.usersByID, id)
	return nil
}

func TestUserUseCase_RegisterUser_Success(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	input := dto.CreateUserInput{
		Name:     "John Doe",
		Email:    "JOHN_DOE@email.com",
		Password: "123",
	}

	u, err := uc.RegisterUser(input)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if u.ID == (pkgentity.ID{}) {
		t.Fatalf("expected non-empty user ID")
	}

	if u.Email != "john_doe@email.com" {
		t.Fatalf("expected normalized email %q, got %q", "john_doe@email.com", u.Email)
	}

	// Senha deve estar hasheada e validável
	if !u.CheckPassword("123") {
		t.Fatalf("expected password to match")
	}
}

func TestUserUseCase_RegisterUser_InvalidEmail(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	_, err := uc.RegisterUser(dto.CreateUserInput{
		Name:     "John",
		Email:    "invalid-email",
		Password: "123",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got: %v", err)
	}
}

func TestUserUseCase_RegisterUser_RepoError(t *testing.T) {
	repo := newFakeUserRepo()
	repo.createErr = errors.New("db error")
	uc := NewUserUseCase(repo)

	_, err := uc.RegisterUser(dto.CreateUserInput{
		Name:     "John",
		Email:    "john@email.com",
		Password: "123",
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "db error" {
		t.Fatalf("expected db error, got: %v", err)
	}
}

func TestUserUseCase_Login_Success(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	u, err := entity.NewUser("John", "john@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("expected no error creating user in repo, got: %v", err)
	}

	// Deve funcionar com email em caixa alta (normalização do usecase).
	userID, err := uc.Login("JOHN@EMAIL.COM", "123")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if userID != u.ID {
		t.Fatalf("expected userID %v, got %v", u.ID, userID)
	}
}

func TestUserUseCase_Login_UserNotFound(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	_, err := uc.Login("missing@email.com", "123")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserUseCase_Login_WrongPassword(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	u, err := entity.NewUser("John", "john@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	_ = repo.Create(u)

	_, err = uc.Login("john@email.com", "wrong")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestUserUseCase_GetMe_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	_, err := uc.GetMe(pkgentity.NewId())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserUseCase_UpdateMe_Success_NameAndPassword(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	u, err := entity.NewUser("John", "john@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	_ = repo.Create(u)

	newName := "John Updated"
	newPassword := "456"

	updated, err := uc.UpdateMe(u.ID, dto.UpdateUserInput{
		Name:     &newName,
		Password: &newPassword,
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if updated.Name != "John Updated" {
		t.Fatalf("expected name %q, got %q", "John Updated", updated.Name)
	}

	if !updated.CheckPassword("456") {
		t.Fatalf("expected updated password to match")
	}
}

func TestUserUseCase_UpdateMe_InvalidName_Empty(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	u, err := entity.NewUser("John", "john@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	_ = repo.Create(u)

	empty := "   "
	_, err = uc.UpdateMe(u.ID, dto.UpdateUserInput{Name: &empty})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, entity.ErrRequired) {
		t.Fatalf("expected ErrRequired, got: %v", err)
	}
}

func TestUserUseCase_UpdateMe_NotFound(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	name := "John"
	_, err := uc.UpdateMe(pkgentity.NewId(), dto.UpdateUserInput{Name: &name})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}

func TestUserUseCase_DeleteMe_Success(t *testing.T) {
	repo := newFakeUserRepo()
	uc := NewUserUseCase(repo)

	u, err := entity.NewUser("John", "john@email.com", "123")
	if err != nil {
		t.Fatalf("expected no error creating entity, got: %v", err)
	}
	_ = repo.Create(u)

	if err := uc.DeleteMe(u.ID); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	_, err = uc.GetMe(u.ID)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
	}
}
