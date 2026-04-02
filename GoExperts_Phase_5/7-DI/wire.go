//go:build wireinject
// +build wireinject

package main

import (
	"database/sql"

	"github.com/google/wire"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/7-DI/product"
)

var setRepositoryDependency = wire.NewSet(
	product.NewProductRepository,
	// wire.Bind: Toda vez que receber uma interface retorne o próprio repositório
	wire.Bind(new(product.ProductRepositoryInterface), new(*product.ProductRepository)),
)

func NewUseCase(db *sql.DB) *product.ProductUseCase {
	wire.Build(
		setRepositoryDependency,
		product.NewProductUseCase,
	)
	return &product.ProductUseCase{}
}
