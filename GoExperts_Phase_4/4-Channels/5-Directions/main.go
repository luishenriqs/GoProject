package main

import (
	"fmt"
)

/*
Resumo — Direcionamento de channels: send-only vs receive-only (papéis explícitos)

Este exemplo demonstra como criar um channel e passá-lo para funções com restrição de direção,
fazendo o compilador garantir que cada função cumpra apenas o seu papel:
- `send`: somente envia dados (produtor)
- `receive`: somente recebe dados (consumidor)

Fluxo de execução
1) `main` cria um channel sem buffer:
   - `ch := make(chan string)` (bidirecional por padrão: pode enviar e receber)
2) `main` inicia uma goroutine produtora:
   - `go send("Hello", ch)` envia a string para o channel
3) `main` executa a função consumidora:
   - `receive(ch)` recebe do channel e imprime a mensagem

Restrição de direção nos parâmetros
- `send(nome string, canal chan<- string)`
  - `chan<- string` significa “somente envio”.
  - Dentro de `send`, o compilador permite `canal <- nome`, mas proíbe leituras (`<-canal`).

- `receive(data <-chan string)`
  - `<-chan string` significa “somente recebimento”.
  - Dentro de `receive`, o compilador permite `<-data`, mas proíbe envios (`data <- ...`).

Sincronização implícita (channel sem buffer)
- Como o channel não tem buffer, existe um “handshake” obrigatório:
  - `send` bloqueia em `canal <- nome` até `receive` fazer a leitura `<-data`.
  - `receive` bloqueia em `<-data` até `send` enviar.
- Isso garante que o programa só imprime depois que a mensagem foi realmente produzida,
  e também evita que a `main` finalize antes da goroutine conseguir entregar o valor.

Objetivo didático
- Mostrar que o mesmo channel (criado bidirecional em `main`) pode ser passado para funções
  como send-only ou receive-only para aumentar clareza e segurança em código concorrente.
*/

func main() {
	ch := make(chan string) // channel bidirecional

	go send("Hello", ch)
	receive(ch)
}

// somente envia dados (produtor) * Envia pro canal
func send(nome string, canal chan<- string) {
	canal <- nome
}

// somente recebe dados (consumidor) * Recebe do canal
func receive(data <-chan string) {
	fmt.Println(<-data)
}
