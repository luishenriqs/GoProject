package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	pkgentity "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/pkg/entity"
	"gorm.io/gorm"
)

// --- Fakes (em memória) para montar UseCase real nos testes ---

type fakeUserRepo struct {
	usersByID    map[pkgentity.ID]*entity.User
	usersByEmail map[string]pkgentity.ID
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		usersByID:    make(map[pkgentity.ID]*entity.User),
		usersByEmail: make(map[string]pkgentity.ID),
	}
}

func (r *fakeUserRepo) Create(user *entity.User) error {
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
	u := r.usersByID[id]
	if u == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) Update(user *entity.User) error {
	if _, ok := r.usersByID[user.ID]; !ok {
		return gorm.ErrRecordNotFound
	}
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user.ID
	return nil
}

func (r *fakeUserRepo) Delete(id pkgentity.ID) error {
	u := r.usersByID[id]
	if u == nil {
		return gorm.ErrRecordNotFound
	}
	delete(r.usersByEmail, u.Email)
	delete(r.usersByID, id)
	return nil
}

// --- Helpers ---

func makeBearerToken(t *testing.T, tokenAuth *jwtauth.JWTAuth, userID pkgentity.ID) string {
	t.Helper()

	exp := time.Now().Add(1 * time.Hour).Unix()
	_, tokenString, err := tokenAuth.Encode(map[string]interface{}{
		"sub": userID.String(),
		"exp": exp,
	})
	if err != nil {
		t.Fatalf("expected no error encoding token, got: %v", err)
	}
	return tokenString
}

func TestUserHandler_CreateUser_BadJSON_400(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	h := NewUserHandler(uc, jwtauth.New("HS256", []byte("secret"), nil), 300)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandler_CreateUser_Success_201(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	h := NewUserHandler(uc, jwtauth.New("HS256", []byte("secret"), nil), 300)

	body, _ := json.Marshal(dto.CreateUserInput{
		Name:     "John Doe",
		Email:    "JOHN@EMAIL.COM",
		Password: "123",
	})

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.CreateUser(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp map[string]interface{}
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if resp["email"] != "john@email.com" {
		t.Fatalf("expected normalized email, got %v", resp["email"])
	}

	if resp["id"] == "" {
		t.Fatalf("expected id to be present")
	}
}

func TestUserHandler_Login_BadJSON_400(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	h := NewUserHandler(uc, jwtauth.New("HS256", []byte("secret"), nil), 300)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandler_Login_WrongCredentials_401(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	h := NewUserHandler(uc, jwtauth.New("HS256", []byte("secret"), nil), 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	body, _ := json.Marshal(dto.LoginInput{
		Email:    "john@email.com",
		Password: "wrong",
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUserHandler_Login_Success_200_ReturnsTokenWithSub(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	body, _ := json.Marshal(dto.LoginInput{
		Email:    "john@email.com",
		Password: "123",
	})

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	var resp map[string]string
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	tokenStr := resp["access_token"]
	if tokenStr == "" {
		t.Fatalf("expected access_token to be present")
	}

	token, err := jwtauth.VerifyToken(tokenAuth, tokenStr)
	if err != nil || token == nil {
		t.Fatalf("expected token to be verifiable, err=%v", err)
	}

	subVal, ok := token.Get("sub")
	if !ok {
		t.Fatalf("expected sub claim")
	}

	sub, _ := subVal.(string)
	if sub != u.ID.String() {
		t.Fatalf("expected sub %q, got %q", u.ID.String(), sub)
	}
}

func TestUserHandler_GetMe_NoAuth_401(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.GetMe))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUserHandler_GetMe_Success_200(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.GetMe))

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerToken(t, tokenAuth, u.ID))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUserHandler_UpdateMe_BadJSON_400(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.UpdateMe))

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString("{invalid"))
	req.Header.Set("Authorization", "Bearer "+makeBearerToken(t, tokenAuth, u.ID))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUserHandler_UpdateMe_Success_200(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	newName := "John Updated"
	body, _ := json.Marshal(dto.UpdateUserInput{
		Name: &newName,
	})

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.UpdateMe))

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+makeBearerToken(t, tokenAuth, u.ID))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rec.Code)
	}

	updated, err := uc.GetMe(u.ID)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if updated.Name != "John Updated" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
}

func TestUserHandler_DeleteMe_Success_204(t *testing.T) {
	repo := newFakeUserRepo()
	uc := usecase.NewUserUseCase(repo)
	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)
	h := NewUserHandler(uc, tokenAuth, 300)

	u, _ := entity.NewUser("John", "john@email.com", "123")
	_ = repo.Create(u)

	protected := middleware.AuthMiddleware(tokenAuth)(http.HandlerFunc(h.DeleteMe))

	req := httptest.NewRequest(http.MethodDelete, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+makeBearerToken(t, tokenAuth, u.ID))
	rec := httptest.NewRecorder()

	protected.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, rec.Code)
	}
}
