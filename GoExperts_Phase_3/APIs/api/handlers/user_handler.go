package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/dto"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
	"gorm.io/gorm"
)

// UserHandler concentra endpoints HTTP relacionados a User/Auth.
type UserHandler struct {
	UseCase      *usecase.UserUseCase
	TokenAuth    *jwtauth.JWTAuth
	JWTExpiresIn int // em segundos
}

func NewUserHandler(uc *usecase.UserUseCase, tokenAuth *jwtauth.JWTAuth, jwtExpiresIn int) *UserHandler {
	return &UserHandler{
		UseCase:      uc,
		TokenAuth:    tokenAuth,
		JWTExpiresIn: jwtExpiresIn,
	}
}

// userResponse é o formato seguro de retorno (sem password).
type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:    u.ID.String(),
		Name:  u.Name,
		Email: u.Email,
	}
}

// CreateUser atende POST /users.
// - body inválido -> 400
// - validação de domínio -> 400
// - erro interno -> 500
// - sucesso -> 201 + JSON (sem password)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input dto.CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.UseCase.RegisterUser(input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) ||
			errors.Is(err, entity.ErrInvalidEmail) ||
			errors.Is(err, entity.ErrWeakPassword) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

// Login atende POST /login.
// - body inválido -> 400
// - credenciais inválidas / usuário não encontrado -> 401
// - erro interno -> 500
// - sucesso -> 200 + JSON {access_token}
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input dto.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, err := h.UseCase.Login(input.Email, input.Password)
	if err != nil {
		// Tratamos not found como 401 para não vazar existência de usuário.
		if errors.Is(err, usecase.ErrInvalidCredentials) || errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	exp := time.Now().Add(time.Duration(h.JWTExpiresIn) * time.Second).Unix()
	_, tokenString, err := h.TokenAuth.Encode(map[string]interface{}{
		"sub": userID.String(),
		"exp": exp,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"access_token": tokenString,
	})
}

// GetMe atende GET /users/me (protegido).
// - sem userID no contexto -> 401
// - not found -> 404
// - sucesso -> 200 + JSON (sem password)
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	u, err := h.UseCase.GetMe(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

// UpdateMe atende PUT /users/me (protegido).
// - sem userID no contexto -> 401
// - body inválido -> 400
// - validação -> 400
// - not found -> 404
// - sucesso -> 200 + JSON (sem password)
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var input dto.UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, err := h.UseCase.UpdateMe(userID, input)
	if err != nil {
		if errors.Is(err, entity.ErrRequired) || errors.Is(err, entity.ErrWeakPassword) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toUserResponse(u))
}

// DeleteMe atende DELETE /users/me (protegido).
// - sem userID no contexto -> 401
// - not found -> 404
// - sucesso -> 204
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.UseCase.DeleteMe(userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
