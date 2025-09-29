package main

import (
	"log"
	"net/http"
)

func main() {
    fileServer := http.FileServer(http.Dir("./public"))
    mux := http.NewServeMux()
    // Criação de end points (/ , /blog)
    mux.Handle("/", fileServer)
    mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello from blog"))
    })
    log.Fatal(http.ListenAndServe(":8080", mux)) // Pacote de logs
}
