package main

/*
Este exemplo demonstra o uso de um **channel bufferizado** (com capacidade) para desacoplar
o envio (send) do recebimento (receive).

O que acontece aqui:
- `ch := make(chan string, 2)` cria um channel de `string` com **buffer de tamanho 2**.
  Isso significa que até 2 valores podem ficar “enfileirados” no channel sem que exista
  um receiver lendo imediatamente.

Envios:
- `ch <- "Hello"`
- `ch <- "World"`
Como o channel tem buffer, esses dois envios **não bloqueiam** (o buffer ainda tem espaço).

Recebimentos:
- `println(<-ch)` lê o primeiro valor enfileirado (ordem FIFO) -> imprime "Hello"
- `println(<-ch)` lê o segundo valor enfileirado -> imprime "World"

Pontos-chave:
- Channels bufferizados são úteis quando você quer que producers enviem “em rajadas” sem
  precisar sincronizar exatamente com o consumidor a cada envio.
- Mesmo sendo bufferizado, a ordem de leitura continua sendo **FIFO** (primeiro que entra,
  primeiro que sai).
- Se você tentasse enviar mais de 2 valores sem consumir nenhum, o 3º envio bloquearia,
  porque o buffer estaria cheio.
*/
func main() {
	ch := make(chan string, 5)
	ch <- "Hello"
	ch <- "World"

	println(<-ch)
	println(<-ch)
}
