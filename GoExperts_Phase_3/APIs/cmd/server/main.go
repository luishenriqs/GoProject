package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/api/handlers"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/configs"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/infra/database"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/entity"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_3/APIs/internal/usecase"
)

func main() {
	cfg, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	db, err := openDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&entity.Product{}, &entity.User{}); err != nil {
		log.Fatal(err)
	}

	userRepo := database.NewUserDb(db)
	productRepo := database.NewProductDB(db)

	userUC := usecase.NewUserUseCase(userRepo)
	productUC := usecase.NewProductUseCase(productRepo)

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

	mux := api.NewMux(cfg.TokenAuth, userRoutes, productRoutes)

	port := cfg.WebServerPort
	if strings.TrimSpace(port) == "" {
		port = "8000"
	}

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}

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
