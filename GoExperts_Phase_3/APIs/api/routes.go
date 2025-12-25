package api

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/middleware"
)

/*
UserRoutes centraliza as funções HTTP (handlers) da entidade User para serem registradas no mux.

Observações:
- Sem router/framework: os próprios handlers devem validar método (POST/GET/PUT/DELETE) e responder 405 quando aplicável.
- A função NewMux apenas organiza o registro e aplica o middleware nas rotas protegidas.
*/
type UserRoutes struct {
	Users http.HandlerFunc // esperado: POST /users (registro)
	Login http.HandlerFunc // esperado: POST /login
	Me    http.HandlerFunc // esperado: GET/PUT/DELETE /users/me
}

/*
ProductRoutes centraliza as funções HTTP (handlers) da entidade Product para serem registradas no mux.

Observações:
- Sem router/framework: para rotas com {id}, registramos o prefixo "/products/" e o handler faz o parse manual do id.
- A função NewMux apenas organiza o registro e aplica o middleware nas rotas protegidas.
*/
type ProductRoutes struct {
	Collection http.HandlerFunc // esperado: POST /products e GET /products (list)
	Item       http.HandlerFunc // esperado: GET/PUT/DELETE /products/{id} via prefixo "/products/"
}

/*
NewMux cria e retorna um *http.ServeMux com todas as rotas da API registradas.

Rotas públicas:
- POST /users
- POST /login

Rotas protegidas (JWT obrigatório):
- GET/PUT/DELETE /users/me
- POST/GET /products
- GET/PUT/DELETE /products/{id}

Importante:
- O agrupamento das rotas protegidas usa o middleware real middleware.AuthMiddleware(tokenAuth).
- Este arquivo não faz wiring de DB/usecases/handlers; isso fica para a Etapa 7 (main.go).
*/
func NewMux(tokenAuth *jwtauth.JWTAuth, user UserRoutes, product ProductRoutes) *http.ServeMux {
	mux := http.NewServeMux()

	// Rotas públicas
	mux.HandleFunc("/users", user.Users)
	mux.HandleFunc("/login", user.Login)

	// Rotas protegidas (JWT)
	protected := middleware.AuthMiddleware(tokenAuth)

	mux.Handle("/users/me", protected(http.HandlerFunc(user.Me)))

	// Collection: /products (POST e GET list)
	mux.Handle("/products", protected(http.HandlerFunc(product.Collection)))

	// Item: /products/{id} (prefixo)
	mux.Handle("/products/", protected(http.HandlerFunc(product.Item)))

	return mux
}
