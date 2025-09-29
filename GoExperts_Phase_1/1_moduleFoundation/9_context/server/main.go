package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", handler) // Cria um servidor
	http.ListenAndServe(":8080", nil) // Cria um servidor
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // r.Context() ==> Context que veio da requisição
	log.Println("Request iniciada")
	defer log.Println("Request finalizada")

	select {
	case <-time.After(9 * time.Second):
		// Imprime no command line stdout
		log.Println("Request processada com sucesso")
		// Imprime no browser
		w.Write([]byte("Request processada com sucesso"))

	case <-ctx.Done():
		// Imprime no command line stdout
		log.Println("Request cancelada pelo cliente")
	}


    /*
        CASO DE USO: case <-ctx.Done():
        O servidor esta processando grande volume de dados
        O cliente cancela no browser
        O servidor interrompe o processamento imediatamente
    */
}
