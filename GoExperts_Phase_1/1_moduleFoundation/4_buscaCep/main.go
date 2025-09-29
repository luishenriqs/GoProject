package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Cria a struct
type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Unidade     string `json:"unidade"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Estado      string `json:"estado"`
	Regiao      string `json:"regiao"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

// Importante!
// Para testar digite no terminal passando o argumento "cep":
// go run main.go 14093070


func main() {
	for _, cep := range os.Args[1:] {
		req, err := http.Get("http://viacep.com.br/ws/" + cep + "/json/") // Executa a requisição com o argumento passado
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer requisição: %v\n", err)
		}
		defer req.Body.Close()
		res, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
		}
		// fmt.Println(string(res)) // Imprime no formato de string

		var data ViaCEP
		err = json.Unmarshal(res, &data) // Parse para o formato da struct
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
		}
		// fmt.Println(data) // Imprime no formato da struct

		// Criação de arquivo
		file, err := os.Create("Endereco.txt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao criar um arquivo: %v\n", err)
		}
		defer file.Close()

		// Escrita
		_, err = file.WriteString(fmt.Sprintf("Localidade: %s", data.Localidade)) // fmt.Sprintf - Não necessita salvar em variável
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao escrever em um arquivo: %v\n", err)
		}

		// Leitura
		arquivo, err := os.ReadFile("Endereco.txt")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(arquivo))
	}
}



// Para abrir o arquivo no terminal digite: cat endereco.txt

// go build -o cep main.go ===> Cria um arquivo cep
// ./cep 14093070 ==> Executa o arquivo passando o parâmetro