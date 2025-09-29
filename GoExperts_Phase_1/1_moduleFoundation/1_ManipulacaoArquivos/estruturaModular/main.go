package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Criação e Escrita
	if err := criarArquivo(); err != nil {
		panic(err)
	}

	// Leitura direta
	if err := leituraSimples(); err != nil {
		panic(err)
	}

	// Leitura com buffer
	if err := leituraComBuffer(); err != nil {
		panic(err)
	}

	// Remoção
	if err := os.Remove("Arquivo.txt"); err != nil {
		panic(err)
	}
}

func criarArquivo() error {
	file, err := os.Create("Arquivo.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	// Escrita
	_, err = file.Write([]byte("Primeira linha do arquivo txt\n"))
	if err != nil {
		return err
	}
	tamanho, err := file.WriteString("Segunda linha do arquivo txt")
	if err != nil {
		return err
	}

	fmt.Printf("Tamanho: %d bites\n", tamanho)
	return nil
}

func leituraSimples() error {
	arquivo, err := os.ReadFile("Arquivo.txt")
	if err != nil {
		return err
	}
	fmt.Println(string(arquivo))
	return nil
}

func leituraComBuffer() error {
	arquivoBuf, err := os.Open("Arquivo.txt")
	if err != nil {
		return err
	}
	defer arquivoBuf.Close()

	reader := bufio.NewReader(arquivoBuf)
	buffer := make([]byte, 8)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			fmt.Println(string(buffer[:n]))
		}
		if err != nil {
			break
		}
	}
	return nil
}
