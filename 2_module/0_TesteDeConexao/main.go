package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Entre nesta pasta e rode: go run main.go
// Entre nesta pasta e rode: go run main.go
// Entre nesta pasta e rode: go run main.go

func main() {
	// DSN: <user>:<password>@tcp(<host>:<port>)/<dbname>
	dsn := "myuser:root@tcp(localhost:3306)/goexpert"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}
	defer db.Close()

	// Testa conexão
	err = db.Ping()
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v", err)
	}

	fmt.Println("Conexão com MySQL estabelecida com sucesso!")
}
