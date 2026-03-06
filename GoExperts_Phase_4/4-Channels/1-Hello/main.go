package main

import "fmt"

/*
Resumo — Channels não bufferizados: sincronização por bloqueio (send/receive)

Este exemplo demonstra como um channel (canal) em Go funciona como mecanismo de
comunicação e sincronização entre goroutines quando ele é criado SEM buffer
(`make(chan string)`).

Ponto principal: channel sem buffer exige “encontro” (handshake)
- Um envio (`canal <- valor`) só completa quando existe um recebimento pronto
  (`<-canal`) ao mesmo tempo.
- Um recebimento (`<-canal`) só completa quando existe um envio pronto.
- Portanto, o channel sem buffer sincroniza automaticamente produtor e consumidor,
  porque um lado bloqueia até o outro lado estar disponível.

O que acontece neste código
1) A goroutine principal (Thread 1) cria `canal := make(chan string)`.
   Nesse momento, o channel está “vazio” e não tem capacidade para armazenar mensagens.

2) Duas goroutines (Thread 2 e outra goroutine) tentam enviar mensagens:
   - Goroutine A: `canal <- "Olá Mundo!"`
   - Goroutine B: `canal <- "Hello World!"`

3) Na Thread 1 existe apenas UM recebimento:
   - `msg := <-canal`
   Isso significa que somente UMA das duas goroutines conseguirá “encontrar” o
   recebimento e completar o envio. A outra ficará bloqueada tentando enviar,
   porque não existe um segundo `<-canal` para consumi-la.

Consequências observáveis
- Apenas UMA mensagem é impressa (a que “ganhou a corrida” e foi recebida primeiro).
- A outra goroutine fica travada no envio, o que dá a impressão de “não funciona”
  ou “não preenche”, mas na verdade está bloqueada esperando um consumidor.
- A ordem não é determinística: pode imprimir "Olá Mundo!" ou "Hello World!"
  dependendo do escalonamento do scheduler.

Lição deste exemplo
- Em channels sem buffer, o número de envios precisa corresponder ao número de
  recebimentos (ou então algum envio ficará bloqueado).
- Para receber ambas as mensagens, seria necessário um segundo `<-canal` (ou outra
  estratégia de consumo), ou então usar um channel com buffer para permitir envios
  sem bloqueio imediato.
*/

// Thread 1
func main() {
	canal := make(chan string) // Vazio

	// Thread 2
	go func() {
		canal <- "Olá Mundo!" // Está cheio
	}()

	go func() {
		canal <- "Hello World!" // Não preenche / Não funciona
	}()

	// Thread 1
	msg := <-canal // Canal esvazia
	fmt.Println(msg)
}
