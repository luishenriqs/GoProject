package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// Product representa um item com ID, nome e preço.
type Product struct {
	ID    string
	Name  string
	Price float64
}

func main() {
	// DSN: <user>:<password>@tcp(<host>:<port>)/<dbname>
	dsn := "myuser:root@tcp(localhost:3306)/goexpert"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}
	defer db.Close()

	// Create
	product := NewProduct("IPhone", 14199.99)
	err = insertProduct(db, product)
	if err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}
	// Update
	product.Name = "IPad"
	product.Price = 4999.00
	err = updateProduct(db, product)
	if err != nil {
		log.Fatalf("Erro ao atualizar produto: %v", err)
	}

	// Select 1 product
	p, err := selectProduct(db, product.ID)
	if err != nil {
		log.Fatalf("Erro ao buscar produto: %v", err)
	}
	if p == nil {
		log.Println("Produto não encontrado.")
	} else {
		// log.Printf("Produto encontrado: ID=%s | Nome=%s | Preço=%.2f\n", p.ID, p.Name, p.Price)
		// log.Printf("O produto %v, possui o preço %.2f\n", p.Name, p.Price)
	}

	// Select all products
	allProducts, err := selectAllProducts(db)
	if err != nil {
		log.Fatalf("Erro ao buscar todos os produtos: %v", err)
	}
	for _, p := range allProducts {
		fmt.Printf("Produto: %v, preço: %2.f\n", p.Name, p.Price)
	}

	// Delete
	// err = deleteProduct(db, p.ID)
	// 	if err != nil {
	// 	log.Fatalf("Erro ao deletar produto: %v", err)
	// }
}

// NewProduct cria uma nova instância de Product com um ID único gerado automaticamente.
// Esta função é responsável apenas por construir um objeto em memória,
// atribuindo um UUID como identificador. Não realiza nenhuma operação de persistência no banco.
//
// Parâmetros:
// - name: string — nome do produto a ser criado.
// - price: float64 — preço do produto.
//
// Retorno:
// - *Product: ponteiro para a instância de Product criada com os dados fornecidos.
func NewProduct(name string, price float64) *Product {
	return &Product{
		ID:    uuid.New().String(),
		Name:  name,
		Price: price,
	}
}

// insertProduct insere um novo produto na tabela "products" do banco de dados MySQL.
// Agora aceita um *Product (ponteiro), permitindo acesso direto e evitando cópias.
// Ideal quando o objeto é criado com NewProduct (que já retorna *Product).
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
// - product: ponteiro para struct contendo ID, nome e preço do produto.
//
// Retorno:
// - error: erro da operação, caso ocorra.
// Utiliza prepared statement com placeholders para proteger contra SQL injection.
// Espera receber uma instância de Product previamente preenchida. (Criada na função NewProduct)
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
// - product: struct contendo ID, nome e preço do produto.
//
// Retorno:
// - error: erro da operação, caso ocorra (preparação ou execução da query); caso contrário, retorna nil.
func insertProduct(db *sql.DB, product *Product) error {

	stmt, err := db.Prepare("INSERT INTO products (id, name, price) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(product.ID, product.Name, product.Price)
	if err != nil {
		return err
	}

	return nil
}

// updateProduct atualiza os campos name e price de um produto existente na tabela "products".
// Utiliza prepared statement com placeholders para proteger contra SQL injection.
// Espera receber uma instância de Product com o ID do produto a ser atualizado.
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
// - product: struct contendo ID, nome e preço atualizados.
//
// Retorno:
// - error: erro da operação, caso ocorra (preparação ou execução da query); caso contrário, retorna nil.
func updateProduct(db *sql.DB, product *Product) error {

	stmt, err := db.Prepare("UPDATE products SET name = ?, price = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(product.Name, product.Price, product.ID)
	if err != nil {
		return err
	}

	return nil
}

// selectProduct busca um produto na tabela "products" usando o ID como filtro.
// Retorna um ponteiro para Product se encontrado, ou erro caso contrário.
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
// - id: identificador único (UUID) do produto.
//
// Retorno:
// - *Product: ponteiro para o produto encontrado (ou nil se não encontrado).
// - error: erro da operação, caso ocorra.
func selectProduct(db *sql.DB, id string) (*Product, error) {
	stmt, err := db.Prepare("SELECT id, name, price FROM products WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var p Product
	err = stmt.QueryRow(id).Scan(&p.ID, &p.Name, &p.Price)
	if err == sql.ErrNoRows {
		return nil, nil // Produto não encontrado, retorna nil sem erro
	}
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// selectAllProducts retorna todos os produtos existentes na tabela "products".
// Executa uma query SELECT fixa (sem parâmetros externos), segura contra SQL injection por definição.
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
//
// Retorno:
// - []Product: slice com todos os produtos encontrados (pode estar vazia).
// - error: erro da operação, caso ocorra.
func selectAllProducts(db *sql.DB) ([]Product, error) {
	query := "SELECT id, name, price FROM products"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

// deleteProduct remove um produto da tabela "products" com base no ID fornecido.
// A operação é segura contra SQL injection via prepared statement.
//
// Parâmetros:
// - db: conexão ativa com o banco de dados (*sql.DB).
// - id: identificador único (UUID) do produto a ser removido.
//
// Retorno:
// - error: erro da operação, caso ocorra.
func deleteProduct(db *sql.DB, id string) error {
	stmt, err := db.Prepare("DELETE FROM products WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id) // Ao invés de buscar (Query) nós executamos (Exec)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Println("Nenhum produto encontrado para deletar.")
	}

	return nil
}
