package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/glebarez/sqlite"
	chimw "github.com/go-chi/chi/v5/middleware"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/handlers"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/configs"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/infra/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"

	_ "github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Go Experts API
// @version         1.0
// @description     API de estudos (Users e Products) usando net/http, GORM e JWT.
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Informe: Bearer {seu_token}

//go:generate swag init -g main.go -o ../../docs -d .,../../api,../../internal/dto,../../internal/entity,../../internal/usecase,../../infra/database,../../pkg/entity,../../configs --parseDependency --parseInternal

/*
main é o ponto de entrada da aplicação HTTP. Ele carrega as configurações, inicializa
a infraestrutura (banco e migrações), monta as dependências (repositórios, use cases e handlers),
configura as rotas HTTP no mux, envolve o mux com middleware de logging e inicia o servidor.

Fluxo:
 1. Carrega as configurações da aplicação via configs.LoadConfig(".").
    - Se falhar, encerra o processo com log.Fatal.
 2. Abre a conexão com o banco de dados via openDatabase(cfg).
    - Se falhar, encerra o processo com log.Fatal.
 3. Executa as migrações do schema via db.AutoMigrate(&entity.Product{}, &entity.User{}).
    - Se falhar, encerra o processo com log.Fatal.
 4. Instancia os repositórios (camada infra/database):
    - userRepo := database.NewUserDb(db)
    - productRepo := database.NewProductDB(db)
 5. Instancia os casos de uso (camada internal/usecase):
    - userUC := usecase.NewUserUseCase(userRepo)
    - productUC := usecase.NewProductUseCase(productRepo)
 6. Instancia os handlers HTTP (camada api/handlers), injetando dependências e configs necessárias:
    - userHandler := handlers.NewUserHandler(userUC, cfg.TokenAuth, cfg.JWTExpiresIn)
    - productHandler := handlers.NewProductHandler(productUC)
 7. Define os contratos de rotas esperados por api.NewMux:
    - userRoutes (POST /users, POST /login, e GET/PUT/DELETE /users/me via multiplexação por método)
    - productRoutes (POST/GET /products e GET/PUT/DELETE /products/{id} via multiplexação por método)
    - Para métodos não suportados, responde com http.StatusMethodNotAllowed.
 8. Cria o mux principal via api.NewMux(cfg.TokenAuth, userRoutes, productRoutes).
    - O mux é responsável por registrar as rotas e aplicar autenticação onde aplicável.
 9. Define a porta do servidor:
    - Usa cfg.WebServerPort; se estiver vazio após trim, usa "8000" como padrão.

10. Envolve o mux com o middleware chimw.Logger(mux) para registrar cada requisição HTTP.
11. Inicia o servidor HTTP com http.ListenAndServe(":"+port, loggedMux).
  - Se falhar, encerra o processo com log.Fatal.

Dependências/efeitos colaterais:
  - Lê configurações do ambiente/arquivos conforme implementação de configs.LoadConfig.
  - Abre conexão com o banco e executa migrações (cria/ajusta tabelas).
  - Registra logs de requisições HTTP (método, path, status, duração) via chimw.Logger.
  - Inicia um servidor HTTP escutando na porta configurada, bloqueando a execução até erro/finalização.

Retorno:
  - Não retorna valor. Em caso de falhas críticas, o processo é encerrado via log.Fatal.
*/
func main() {
	// Carrega as configurações da aplicação via configs.LoadConfig(".")
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	// Abre a conexão com o banco de dados via openDatabase(cfg)
	db, err := openDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Executa as migrações do schema via db.AutoMigrate
	if err := db.AutoMigrate(&entity.Product{}, &entity.User{}); err != nil {
		log.Fatal(err)
	}

	// Instancia os repositórios (camada infra/database)
	userRepo := database.NewUserDb(db)
	productRepo := database.NewProductDB(db)

	// Instancia os casos de uso (camada internal/usecase)
	userUC := usecase.NewUserUseCase(userRepo)
	productUC := usecase.NewProductUseCase(productRepo)

	// Instancia os handlers HTTP (camada api/handlers)
	userHandler := handlers.NewUserHandler(userUC, cfg.TokenAuth, cfg.JWTExpiresIn)
	productHandler := handlers.NewProductHandler(productUC)

	// Monta os contratos esperados por api.NewMux (UserRoutes/ProductRoutes).
	userRoutes := api.UserRoutes{
		Users: userHandler.CreateUser, // POST /users
		Login: userHandler.Login,      // POST /login
		Me: func(w http.ResponseWriter, r *http.Request) {
			// GET/PUT/DELETE /users/me
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
			// POST/GET /products
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
			// GET/PUT/DELETE /products/{id}
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

	// Cria o mux principal via api.NewMux
	mux := api.NewMux(cfg.TokenAuth, userRoutes, productRoutes)

	// Registra a rota do Swagger no mux
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	// Envolve o mux com o middleware chimw.Logger(mux) para registrar cada requisição HTTP.
	loggedMux := chimw.Logger(mux)

	// Define a porta do servidor
	port := cfg.WebServerPort
	if strings.TrimSpace(port) == "" {
		port = "8000"
	}

	log.Printf("listening on :%s", port)

	// Inicia o servidor HTTP com http.ListenAndServe
	if err := http.ListenAndServe(":"+port, loggedMux); err != nil {
		log.Fatal(err)
	}
}

/*
openDatabase abre uma conexão GORM com o banco configurado em cfg, escolhendo o driver
a partir de cfg.DBDriver e montando os parâmetros necessários para cada backend.

Fluxo:
 1. Normaliza cfg.DBDriver com strings.TrimSpace e strings.ToLower.
 2. Seleciona o driver:
    - "mysql":
    a) Monta a DSN com cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort e cfg.DBName,
    incluindo parâmetros de conexão (charset=utf8mb4, parseTime=True, loc=Local).
    b) Abre a conexão via gorm.Open(mysql.Open(dsn), ...).
    - "sqlite" ou "" (vazio):
    a) Abre um SQLite local em arquivo usando sqlite.Open("test.db").
    - Qualquer outro valor:
    a) Retorna erro indicando driver não suportado, incluindo o valor original cfg.DBDriver.
 3. Retorna a instância *gorm.DB ou um erro.

Parâmetros:
  - cfg: ponteiro para configs.Config contendo os campos de configuração do banco, como:
  - DBDriver, DBUser, DBPassword, DBHost, DBPort, DBName.

Retorno:
  - (*gorm.DB, nil) em caso de sucesso.
  - (nil, err) se ocorrer falha ao abrir o banco via GORM.
  - (nil, fmt.Errorf(...)) se cfg.DBDriver indicar um driver não suportado.
*/
func openDatabase(cfg *configs.Config) (*gorm.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.DBDriver))

	switch driver {
	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)
		return gorm.Open(mysql.Open(dsn), &gorm.Config{})

	case "sqlite", "":
		return gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER: %q", cfg.DBDriver)
	}
}
