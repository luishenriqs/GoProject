package main

import (
	"io"
	"net/http"
)

func main() {
    request := http.Client{} // Esse é o Client

    req, err := http.NewRequest("GET", "http://google.com", nil) // Esse é a Requisição
    if err != nil {
        panic(err)
    }

    req.Header.Set("Accept", "application/json") // Configuro/Adiciono informações na req

    resp, err := request.Do(req) // Une o Client com a Requisição --> Client => do => req
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    println(string(body))
}
