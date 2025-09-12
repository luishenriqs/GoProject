package main

import (
	"context"
	"fmt"
)

// tipo customizado para evitar colisão
// Evita esse erro: should not use built-in type string as key for value;
type ctxKey string

func main() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKey("token"), "senha")
	bookHotel(ctx)
}

func bookHotel(ctx context.Context) {
	token := ctx.Value(ctxKey("token"))
	fmt.Println(token)
}