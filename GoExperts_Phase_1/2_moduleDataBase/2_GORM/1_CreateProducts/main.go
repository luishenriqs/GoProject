package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Category struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

// Product representa um item com ID, nome e preço.
// Com GORM, adicionamos tags para mapeamento da tabela e colunas.
type Product struct {
	gorm.Model           // Gera e gerencia automaticamente "createdAt", "updatedAt" e "deletedAt"
	Name         string  `gorm:"type:varchar(100);not null"`
	Price        float64 `gorm:"type:decimal(10,2);not null"`
	CategoryID   uint
	Category     Category
	SerialNumber SerialNumber `gorm:"constraint:OnDelete:CASCADE;"`
}

type SerialNumber struct {
	ID        string `gorm:"type:char(36);primaryKey"`
	Number    string
	ProductID uint
}

func main() {
	// DSN: <user>:<password>@tcp(<host>:<port>)/<dbname>?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "myuser:root@tcp(localhost:3306)/goexpert?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}

	// AutoMigrate garante que a tabela exista e tenha as colunas corretas
	// OBS: migrar na ordem referenciada → dependências depois
	if err := db.AutoMigrate(&Category{}, &Product{}, &SerialNumber{}); err != nil {
		log.Fatalf("Erro ao migrar tabela: %v", err)
	}

	// Cria e insere 1 categoria
	category := NewCategory("Ferramentas")
	if err := insertCategory(db, category); err != nil {
		log.Fatalf("Erro ao inserir categoria: %v", err)
	}

	// Cria e insere 1 produto
	product := NewProduct("Machado", 2989.90, 2)
	if err := insertProduct(db, product); err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}

	// Cria e insere 1 serialNumber
	serialNumber := NewSerialNumber("123456", product.ID)
	if err := insertSerialNumber(db, serialNumber); err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}

	// Cria e insere vários produtos
	// prodList := NewProducts([]struct {
	// 	Name  string
	// 	Price float64
	// }{
	// 	{"Caneta Esferográfica", 8},
	// 	{"Grampeador Compacto", 18.90},
	// 	{"Pacote de Papel A4", 24.99},
	// })
	// if err := insertProducts(db, prodList); err != nil {
	// 	log.Fatalf("Erro ao inserir lista de produtos: %v", err)
	// }

	// Seleciona e imprime categorias
	// allCategories, err := selectAllCategories(db)
	// if err != nil {
	// 	log.Fatalf("Erro ao buscar todas as categorias: %v", err)
	// }
	// for _, c := range allCategories {
	// 	fmt.Printf("Categoria: %v, \n", c.Name)
	// }

	// Seleciona e imprime produtos
	// allProducts, err := selectAllProducts(db)
	// if err != nil {
	// 	log.Fatalf("Erro ao buscar todos os produtos: %v", err)
	// }
	// for _, p := range allProducts {
	// 	fmt.Printf("Produto: %v, preço: %.2f\n", p.Name, p.Price)
	// }

	// Busca produtos e Categoria (Busca relacionada)
	allProductsData, err := selectProductsData(db)
	if err != nil {
		log.Fatalf("Erro ao buscar todos os produtos: %v", err)
	}
	for _, p := range allProductsData {
		fmt.Printf("Produto: %v, preço: %.2f, categoria: %v,  serialNumber: %v,\n", p.Name, p.Price, p.Category, p.SerialNumber.Number)
	}
}

//------------------------------------------------------------------------------

// ########################## CRIA 1 CATEGORIA POR VEZ #########################
// NewCategory cria uma nova instância de Category
func NewCategory(name string) *Category {
	return &Category{
		Name: name,
	}
}

// insertCategory insere uma nova categoria usando GORM
func insertCategory(db *gorm.DB, category *Category) error {
	return db.Create(category).Error
}

// ############################# CRIA SERIAL NUMBER ############################

// NewSerialNumber cria uma nova instância de SerialNumber
func NewSerialNumber(number string, productID uint) *SerialNumber {
	return &SerialNumber{
		ID:        uuid.New().String(),
		Number:    number,
		ProductID: productID,
	}
}

// insertSerialNumber insere um novo serialNumber usando GORM
func insertSerialNumber(db *gorm.DB, serialNumber *SerialNumber) error {
	return db.Create(serialNumber).Error
}

// ########################### CRIA 1 PRODUTO POR VEZ ##########################

// NewProduct cria uma nova instância de Product
func NewProduct(name string, price float64, categoryId uint) *Product {
	return &Product{
		Name:       name,
		Price:      price,
		CategoryID: categoryId,
	}
}

// insertProduct insere um novo produto usando GORM
func insertProduct(db *gorm.DB, product *Product) error {
	return db.Create(product).Error
}

// ######################## CRIA VÁRIOS PRODUTOS POR VEZ #######################

// NewProducts cria várias instâncias de Products
func NewProducts(items []struct {
	Name     string
	Price    float64
	category Category
}) []*Product {
	products := make([]*Product, 0, len(items))
	for _, item := range items {
		products = append(products, &Product{
			Name:       item.Name,
			Price:      item.Price,
			CategoryID: item.category.ID,
		})
	}
	return products
}

// insertProducts insere vários produtos usando GORM
func insertProducts(db *gorm.DB, products []*Product) error {
	return db.Create(&products).Error
}

//------------------------------------------------------------------------------

// ######################## SELECIONA E IMPRIME CATEGORIAS #######################

// selectAllCategories retorna todos as categorias usando GORM
func selectAllCategories(db *gorm.DB) ([]Category, error) {
	var category []Category
	if err := db.Find(&category).Error; err != nil {
		return nil, err
	}
	return category, nil
}

// ######################## SELECIONA E IMPRIME PRODUTOS #######################

// selectAllProducts retorna todos os produtos usando GORM
func selectAllProducts(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

// ############### SELECIONA E IMPRIME PRODUTOS E RELACIONAMENTOS ##############
// selectAllProducts retorna os produtos e seus relacionamentos
func selectProductsData(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").Preload("SerialNumber").Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
