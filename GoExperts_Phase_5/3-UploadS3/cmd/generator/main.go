package main

import (
	"fmt"
	"os"
)

/*
main cria 1000 arquivos de texto no diretório ./tmp, nomeados de forma
sequencial no formato file0.txt até file999.txt, e escreve o conteúdo
"Hello, World!" em cada um deles.

Parâmetros:
- nenhum.

Retorno:
- nenhum.
*/
func main() {
	for i := 0; i < 1000; i++ {
		f, err := os.Create(fmt.Sprintf("./tmp/file%d.txt", i))
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString("Hello, World!")
		if err != nil {
			f.Close()
			panic(err)
		}

		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
}
