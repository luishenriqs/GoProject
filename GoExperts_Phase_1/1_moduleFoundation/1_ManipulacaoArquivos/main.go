package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Criação
	file, err := os.Create("Arquivo.txt") // Aqui crio e abro o "file"
	if err != nil {
		panic(err)
	}

	// Escrita de slice de bytes - Não se sabe o conteúdo
	file.Write([]byte("Primeira linha do arquivo txt\n"))

	// Escrita quando se sabe ser string
	tamanho, err := file.WriteString("Segunda linha do arquivo txt")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tamanho: %d bites\n", tamanho)
	
	// Leitura
	arquivo, err := os.ReadFile("Arquivo.txt")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(arquivo))

	file.Close() // Aqui fecho o "file"

	// Leitura com Buffer
	arquivoBuf, err := os.Open("Arquivo.txt") // Aqui abro o "arquivoBuf"
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(arquivoBuf)
	buffer := make([]byte, 8)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			break
		}
		fmt.Println(string(buffer[:n]))
	}

	arquivoBuf.Close() // Aqui fecho o "arquivoBuf"

	// Remoção de arquivo
	err = os.Remove("Arquivo.txt")
	if err != nil {
		panic(err)
	}

}