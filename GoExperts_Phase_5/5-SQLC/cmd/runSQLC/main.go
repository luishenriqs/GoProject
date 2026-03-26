package main

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/luishenriqs/GoProject/GoExperts_Phase_5/5-SQLC/internal/db"
)

func main() {
	ctx := context.Background()
	dbConn, err := sql.Open("mysql", "root:root@tcp(localhost:3306)/courses")
	if err != nil {
		panic(err)
	}
	defer dbConn.Close()

	queries := db.New(dbConn)

	// *** CREATE ***

	err = queries.CreateCategory(ctx, db.CreateCategoryParams{
		ID:          uuid.New().String(),
		Name:        "Backend",
		Description: sql.NullString{String: "Backend description", Valid: true},
	})

	if err != nil {
		panic(err)
	}

	// *** UPDATE ***

	// err = queries.UpdateCategory(ctx, db.UpdateCategoryParams{
	// 	ID:          "6c387c80-3918-4e2a-991c-fffab62c2812",
	// 	Name:        "Backend updated",
	// 	Description: sql.NullString{String: "Backend description updated", Valid: true},
	// })

	// *** DELETE ***

	// err = queries.DeleteCategory(ctx, "6c387c80-3918-4e2a-991c-fffab62c2812")
	// if err != nil {
	// 	panic(err)
	// }

	// *** RESULT ***

	categories, err := queries.ListCategories(ctx)
	if err != nil {
		panic(err)
	}

	for _, category := range categories {
		println("CATEGORIES: ", category.ID, category.Name, category.Description.String)
	}

}
