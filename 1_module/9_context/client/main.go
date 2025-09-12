package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	// Neste contexto o client aguarda 10 segundos, se não tiver resposta cancela a chamada
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// CLIENT - Requisição feita para o servidor da porta 8080
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	io.Copy(os.Stdout, res.Body)
}
