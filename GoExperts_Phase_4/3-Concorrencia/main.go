package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var number int64 = 0

/*
Resumo — Corrigindo Race Condition em Go: Mutex vs Atomic

Este arquivo registra duas abordagens clássicas para eliminar race condition ao
acessar e atualizar um contador compartilhado (ex.: `number`) em um servidor HTTP
com múltiplas requisições concorrentes.

Contexto do problema (race condition)
  - Em um handler HTTP, várias goroutines podem executar simultaneamente.
  - A operação `number++` NÃO é atômica: ela envolve (1) ler, (2) somar, (3) escrever.
  - Sob concorrência, duas goroutines podem ler o mesmo valor e sobrescrever a atualização,
    gerando perdas de incremento, valores repetidos e avisos do race detector (`-race`).

Opção 1 — Mutex (sync.Mutex) [abordagem comentada no código]
  - Ideia: usar exclusão mútua para garantir que apenas UMA goroutine por vez
    execute a seção crítica (leitura + incremento + escrita do contador).
  - Como funciona:
  - `mutex.Lock()` antes de acessar/alterar o contador.
  - `mutex.Unlock()` após concluir a atualização.
  - Características:
  - Solução geral e simples de raciocinar: funciona para qualquer trecho crítico,
    inclusive quando há múltiplas variáveis e lógica mais complexa.
  - Introduz bloqueio: goroutines concorrentes esperam na fila, o que pode reduzir
    throughput sob alta concorrência, dependendo do tempo dentro da seção crítica.
  - É a escolha natural quando você precisa proteger um “bloco de operações”
    (mais do que um simples incremento).

Opção 2 — Atomic (sync/atomic) [abordagem ativa no código]
  - Ideia: realizar o incremento como uma operação atômica no nível de instrução/suporte
    do runtime, evitando a janela de corrida existente em `number++`.
  - Como funciona:
  - `atomic.AddUint64(&number, 1)` executa o “ler+somar+escrever” como uma única
    operação atômica, segura sob concorrência.
  - Características:
  - Muito eficiente para contadores e atualizações simples.
  - Evita lock e reduz contenção, geralmente escalando melhor sob carga.
  - Limitação: é apropriado para operações simples em uma variável; para invariantes
    envolvendo múltiplos campos/etapas, Mutex (ou outras técnicas) é mais adequado.

Quando usar cada uma (regra prática)
- Use Atomic quando:
  - você precisa apenas de operações simples (ex.: contador, flags, métricas),
  - e quer minimizar overhead de sincronização.

- Use Mutex quando:
  - a seção crítica envolve múltiplas operações/variáveis,
  - você precisa garantir consistência de um conjunto de passos (invariantes),
  - e a clareza do “bloco protegido” ajuda na manutenção.

Validação recomendada (para confirmar ausência de race)
- Rodar com race detector:
  - `CGO_ENABLED=1 go run -race main.go`
  - e gerar concorrência (ex.: `ab -l -k -n 20000 -c 200 http://localhost:3000/`)

- Resultado esperado após aplicar qualquer uma das duas abordagens:
  - ausência de `WARNING: DATA RACE` no output do servidor.
*/
func main() {

	// m := sync.Mutex{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// m.Lock()
		// number++
		// m.Unlock()
		atomic.AddInt64(&number, 1)
		time.Sleep(200 * time.Millisecond)
		fmt.Fprintf(w, "Você é o visitante número %d", number)
	})

	http.ListenAndServe(":3000", nil)
}
