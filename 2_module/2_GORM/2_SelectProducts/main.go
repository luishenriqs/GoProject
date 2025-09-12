package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Product representa um item com ID, nome e preço.
// Com GORM, adicionamos tags para mapeamento da tabela e colunas.
type Product struct {
	ID         string  `gorm:"type:char(36);primaryKey"`
	Name       string  `gorm:"type:varchar(100);not null"`
	Price      float64 `gorm:"type:decimal(10,2);not null"`
	gorm.Model         // Gera e gerencia automaticamente "createdAt", "updatedAt" e "deletedAt"
}

func main() {
	// DSN: <user>:<password>@tcp(<host>:<port>)/<dbname>?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "myuser:root@tcp(localhost:3306)/goexpert?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}

	// AutoMigrate garante que a tabela exista e tenha as colunas corretas
	if err := db.AutoMigrate(&Product{}); err != nil {
		log.Fatalf("Erro ao migrar tabela: %v", err)
	}

	// New product to be updated
	oldProduct := &Product{
		ID:    uuid.New().String(),
		Name:  "Produto Antigo",
		Price: 150,
	}
	fmt.Printf("Produto: %v, Preço: %.2f\n", oldProduct.Name, oldProduct.Price)
	err = db.Create(oldProduct).Error
	if err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}

	// Update
	oldProduct.Name = "Produto Atualizado"
	oldProduct.Price = 299.00
	if err := updateProduct(db, oldProduct); err != nil {
		log.Fatalf("Erro ao atualizar produto: %v", err)
	}

	// Select 1 product
	p, err := selectProduct(db, oldProduct.ID)
	if err != nil {
		log.Fatalf("Erro ao buscar produto: %v", err)
	}
	if p == nil {
		log.Println("Produto não encontrado.")
	} else {
		fmt.Printf("Produto: %v, Preço: %.2f\n", p.Name, p.Price)
	}

	// Product to be deleted
	toBeDeleted := &Product{
		ID:    uuid.New().String(),
		Name:  "Produto Inútil",
		Price: 1000,
	}
	err = db.Create(toBeDeleted).Error
	if err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}

	// Seleciona e imprime produtos
	allProducts, err := selectAllProducts(db)
	if err != nil {
		log.Fatalf("Erro ao buscar todos os produtos: %v", err)
	}
	for _, p := range allProducts {
		fmt.Printf("Produto: %v, preço: %.2f\n", p.Name, p.Price)
	}

	// Delete
	if err := deleteProduct(db, toBeDeleted.ID); err != nil {
		log.Fatalf("Erro ao deletar produto: %v", err)
	}
}

// selectProduct busca um produto pelo ID usando GORM
func selectProduct(db *gorm.DB, id string) (*Product, error) {
	var p Product
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// selectAllProducts retorna todos os produtos usando GORM
func selectAllProducts(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// updateProduct atualiza um produto usando GORM
func updateProduct(db *gorm.DB, product *Product) error {
	return db.Save(product).Error
}

// deleteProduct remove um produto pelo ID usando GORM
func deleteProduct(db *gorm.DB, id string) error {
	res := db.Delete(&Product{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		log.Println("Nenhum produto encontrado para deletar.")
	}
	return nil
}
