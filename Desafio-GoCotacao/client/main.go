// client.go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

/*
	Visão geral
		Faz GET para http://localhost:8080/cotacao com timeout de 300ms (via context).
		Espera um JSON mínimo {"bid":"<string>"}.
		Valida, e grava em client/cotacao.txt no formato Dólar: {valor}.
		Qualquer falha relevante encerra com log.Fatalf (processo sai com erro).
*/

const (
	serverURL = "http://localhost:8080/cotacao" // serverURL aponta para o endpoint do servidor.
	outPath   = "cotacao.txt"                   // outPath define onde gravar o arquivo (diretório atual do processo → a pasta client/ quando você roda go run main.go dentro dela).
)

// Estrutura mínima compatível com a resposta do servidor; mantém acoplamento baixo.
type bidResponse struct {
	Bid string `json:"bid"`
}

// requestQuote(ctx) — Requisição com timeout de 300ms
func requestQuote(ctx context.Context) (string, error) {
	// Define o SLA do cliente: tudo precisa caber em 300ms.
	// Esse ctx é acoplado ao transporte HTTP; se estourar, a conexão é cancelada.
	ctxReq, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	//NewRequestWithContext garante que cancelamentos/timeout fechem a conexão.
	req, err := http.NewRequestWithContext(ctxReq, http.MethodGet, serverURL, nil)
	if err != nil {
		return "", err
	}

	// Usa o DefaultClient (sem timeout global), deixando o context governar o tempo.
	resp, err := http.DefaultClient.Do(req)
	// Se o servidor demorar além do deadline do cliente → err com context.DeadlineExceeded.
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("[client] timeout (300ms) aguardando server.go: %v", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	// Se o servidor respondeu 502 (falha de upstream) ou 504 (timeout do upstream), isso cai aqui.
	// Resultado: cliente falha e não grava o arquivo (comportamento correto para o desafio).
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status inesperado do servidor: %d", resp.StatusCode)
	}

	// Decodificação streaming (não carrega tudo em memória). Em seguida, valida br.Bid != "".
	var br bidResponse
	if err := json.NewDecoder(resp.Body).Decode(&br); err != nil {
		return "", err
	}
	if br.Bid == "" {
		return "", errors.New("campo bid vazio na resposta do servidor")
	}

	// Se tudo ok, retorna a string do bid para o main().
	return br.Bid, nil
}

/*
Gera exatamente o conteúdo exigido.

	0644 define permissões UNIX-like; no Windows o mode é parcialmente ignorado.
	O arquivo é criado/sobrescrito atômica e sincronicamente do ponto de vista
	do processo (não há flush explícito; suficiente aqui).
*/
func writeQuoteToFile(path, bid string) error {
	content := fmt.Sprintf("Dólar: %s", bid)
	return os.WriteFile(path, []byte(content), 0644)
}

func main() {
	// Caminho feliz: requisita, recebe bid, grava, loga sucesso.
	bid, err := requestQuote(context.Background())

	// Falhas: qualquer erro crítico finaliza o processo com código ≠ 0 (como se espera em um utilitário CLI).
	if err != nil {
		log.Fatalf("[client] erro ao obter cotação: %v", err)
	}

	if err := writeQuoteToFile(outPath, bid); err != nil {
		log.Fatalf("[client] erro ao salvar arquivo: %v", err)
	}

	log.Printf("[client] cotação salva em %s", outPath)
}

/*
	Comportamentos por cenário (relevantes)
		Servidor OK (≤300ms) → cliente recebe 200 OK, grava cotacao.txt.
		Servidor responde 504 rapidamente (upstream do servidor estourou 200ms)
		 → cliente não estoura seu timeout, mas falha por status inesperado.
		Servidor demora >300ms (rede/latência local) → cliente estoura timeout
		e encerra com DeadlineExceeded.
		Servidor caiu/porta errada → erro de conexão; cliente encerra com falha.

	Onde fica o arquivo?
		client/cotacao.txt (porque você roda o cliente dentro da pasta client/).
		Conteúdo: Dólar: <valor> (sem quebra de linha ao final — intencional).
*/
