package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-chi/jwtauth/v5"
	"gorm.io/gorm"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/handlers"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/infra/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
)

type loginResponse struct {
	AccessToken string `json:"access_token"`
}

type userResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type listProductsResponse struct {
	Items []map[string]interface{} `json:"items"`
	Total int64                    `json:"total"`
}

func TestE2E_UserRegisterLoginAndProtectedRoutes(t *testing.T) {
	// 1) Infra (SQLite) + migrations
	db := mustOpenTestDB(t)
	if err := db.AutoMigrate(&entity.Product{}, &entity.User{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	// 2) Repos + usecases
	userRepo := database.NewUserDb(db)
	productRepo := database.NewProductDB(db)

	userUC := usecase.NewUserUseCase(userRepo)
	productUC := usecase.NewProductUseCase(productRepo)

	// 3) TokenAuth + handlers
	tokenAuth := jwtauth.New("HS256", []byte("test-secret"), nil)
	jwtExpiresIn := 3600

	userHandler := handlers.NewUserHandler(userUC, tokenAuth, jwtExpiresIn)
	productHandler := handlers.NewProductHandler(productUC)

	// 4) Rotas (mesmo contrato esperado pelo api.NewMux)
	userRoutes := api.UserRoutes{
		Users: userHandler.CreateUser, // POST /users
		Login: userHandler.Login,      // POST /login
		Me: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				userHandler.GetMe(w, r)
				return
			case http.MethodPut:
				userHandler.UpdateMe(w, r)
				return
			case http.MethodDelete:
				userHandler.DeleteMe(w, r)
				return
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		},
	}

	productRoutes := api.ProductRoutes{
		Collection: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				productHandler.CreateProduct(w, r)
				return
			case http.MethodGet:
				productHandler.ListProducts(w, r)
				return
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		},
		Item: func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				productHandler.GetProductByID(w, r)
				return
			case http.MethodPut:
				productHandler.UpdateProduct(w, r)
				return
			case http.MethodDelete:
				productHandler.DeleteProduct(w, r)
				return
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
		},
	}

	mux := api.NewMux(tokenAuth, userRoutes, productRoutes)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// 5) Garante que rota protegida sem token retorna 401
	{
		resp := mustDo(t, http.MethodGet, srv.URL+"/users/me", nil, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401 on GET /users/me without token, got %d", resp.StatusCode)
		}
	}

	// 6) POST /users (register)
	email := fmt.Sprintf("e2e_%d@test.com", time.Now().UnixNano())
	registerBody := map[string]interface{}{
		"name":     "E2E User",
		"email":    email,
		"password": "StrongPass123!",
	}

	var created userResponse
	{
		resp := mustDo(t, http.MethodPost, srv.URL+"/users", registerBody, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 on POST /users, got %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode POST /users response: %v", err)
		}

		if created.ID == "" || created.Email != email {
			t.Fatalf("unexpected user response: %+v", created)
		}
	}

	// 7) POST /login
	var login loginResponse
	{
		loginBody := map[string]interface{}{
			"email":    email,
			"password": "StrongPass123!",
		}

		resp := mustDo(t, http.MethodPost, srv.URL+"/login", loginBody, nil)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 on POST /login, got %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(&login); err != nil {
			t.Fatalf("decode POST /login response: %v", err)
		}

		if login.AccessToken == "" {
			t.Fatalf("expected access_token, got empty")
		}
	}

	authHeaders := map[string]string{
		"Authorization": "Bearer " + login.AccessToken,
	}

	// 8) GET /users/me com token
	{
		resp := mustDo(t, http.MethodGet, srv.URL+"/users/me", nil, authHeaders)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 on GET /users/me, got %d", resp.StatusCode)
		}

		var me userResponse
		if err := json.NewDecoder(resp.Body).Decode(&me); err != nil {
			t.Fatalf("decode GET /users/me response: %v", err)
		}

		if me.ID != created.ID || me.Email != created.Email {
			t.Fatalf("unexpected /users/me response: %+v", me)
		}
	}

	// 9) POST /products (protegido) + GET /products (list)
	{
		createProductBody := map[string]interface{}{
			"name":  "Product E2E",
			"price": 12.34,
		}

		resp := mustDo(t, http.MethodPost, srv.URL+"/products", createProductBody, authHeaders)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201 on POST /products, got %d", resp.StatusCode)
		}
	}

	{
		resp := mustDo(t, http.MethodGet, srv.URL+"/products?page=1&limit=10&sort=asc", nil, authHeaders)
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 on GET /products, got %d", resp.StatusCode)
		}

		var list listProductsResponse
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode GET /products response: %v", err)
		}

		if list.Total < 1 {
			t.Fatalf("expected total >= 1, got %d", list.Total)
		}
		if len(list.Items) < 1 {
			t.Fatalf("expected items >= 1, got %d", len(list.Items))
		}
	}
}

func mustOpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// SQLite in-memory compartilhado (compatível para uso em múltiplas conexões do gorm).
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test db: %v", err)
	}
	return db
}

func mustDo(t *testing.T, method, url string, body interface{}, headers map[string]string) *http.Response {
	t.Helper()

	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}
