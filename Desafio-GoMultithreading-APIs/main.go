package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Address struct {
	CEP          string
	State        string
	City         string
	Neighborhood string
	Street       string
	Source       string
}

type result struct {
	addr Address
	err  error
}

var cepRegex = regexp.MustCompile(`^\d{8}$`)

/*
DESAFIO — Multithreading e APIs (Go Expert / Full Cycle)
Arquivo: main.go

Este programa resolve o desafio de concorrência fazendo **duas requisições HTTP em paralelo**
para consultar um endereço por CEP e **aceitando a primeira resposta válida** que chegar,
descartando a mais lenta, com **timeout global de 1 segundo**.

✅ Requisitos atendidos
- Disparar simultaneamente duas chamadas:
  - BrasilAPI: https://brasilapi.com.br/api/cep/v1/{cep}
  - ViaCEP:   http://viacep.com.br/ws/{cep}/json/

- Escolher a **API mais rápida** (primeiro resultado válido) e ignorar a mais lenta.
- Exibir no terminal os dados do endereço e qual API venceu.
- Aplicar **timeout de 1s**; se nenhuma responder a tempo, mostrar erro de timeout.

Visão geral do fluxo
1) Validação do input:
  - Lê o CEP via argumento de linha de comando (os.Args[1]).
  - Valida formato com regex `^\d{8}$` (exige 8 dígitos).

2) Controle de tempo e cancelamento:
  - Cria um `context.WithTimeout(..., 1*time.Second)` para impor limite global.
  - Mantém `cancel()` para interromper a request mais lenta assim que houver um vencedor.

3) Concorrência e sincronização:
  - Cria `resCh := make(chan result, 2)` (buffer 2) para receber o retorno das duas goroutines
    sem risco de bloqueio no envio.
  - Dispara `fetchBrasilAPI` e `fetchViaCEP` em goroutines, ambas recebendo o mesmo `ctx`.

4) Escolha da resposta vencedora:
  - O loop `for range 2` aguarda até dois retornos.
  - A cada chegada:
  - Se `err == nil`, a resposta é considerada válida:
  - chama `cancel()` para cancelar a outra chamada em andamento;
  - imprime o endereço via `printAddress` e encerra.
  - Se houver erro, registra e continua aguardando a outra resposta (desde que ainda dentro do timeout).

5) Tratamento de falhas:
  - Se `ctx.Done()` disparar antes de uma resposta válida: imprime timeout (1s).
  - Se as duas APIs retornarem erro dentro do tempo: imprime mensagem geral e detalha erros individuais.

Mapeamento de respostas (normalização)
- BrasilAPI retorna campos `state`, `city`, `neighborhood`, `street`.
- ViaCEP retorna `uf`, `localidade`, `bairro`, `logradouro` e um flag `erro` (CEP inexistente).
O programa normaliza ambos para a struct `Address` e preenche `Source` com o nome da API vencedora.

Observações importantes
  - O cancelamento é efetivo porque as requisições usam `http.NewRequestWithContext`, então
    quando `cancel()` é chamado a request concorrente tende a ser abortada.
  - A ordem de chegada das respostas não é determinística; a seleção depende de latência/rede.
*/
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run . 01153000")
		return
	}

	cep := os.Args[1]
	if !cepRegex.MatchString(cep) {
		fmt.Println("Erro: CEP inválido. Use apenas 8 dígitos (ex: 01153000).")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resCh := make(chan result, 2)

	go fetchBrasilAPI(ctx, cep, resCh)
	go fetchViaCEP(ctx, cep, resCh)

	// Queremos a resposta válida mais rápida.
	// Se a primeira que chegar der erro, ainda tentamos a outra dentro do timeout.
	var errs []error
	for range 2 {
		select {
		case r := <-resCh:
			if r.err == nil {
				// Cancela a outra request (descarta a mais lenta).
				cancel()
				printAddress(r.addr)
				return
			}
			errs = append(errs, r.err)

		case <-ctx.Done():
			fmt.Println("Erro: timeout (1s) — nenhuma API respondeu a tempo.")
			return
		}
	}

	// Se chegou aqui, as duas falharam dentro do tempo.
	fmt.Println("Erro: nenhuma API retornou um resultado válido.")
	for _, e := range errs {
		fmt.Printf("- %v\n", e)
	}
}

func printAddress(a Address) {
	fmt.Printf("API vencedora: %s\n", a.Source)
	fmt.Printf("CEP: %s\n", a.CEP)
	fmt.Printf("Estado: %s\n", a.State)
	fmt.Printf("Cidade: %s\n", a.City)
	fmt.Printf("Bairro: %s\n", a.Neighborhood)
	fmt.Printf("Rua: %s\n", a.Street)
}

func fetchBrasilAPI(ctx context.Context, cep string, ch chan<- result) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		ch <- result{err: fmt.Errorf("BrasilAPI: falha ao criar request: %w", err)}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- result{err: fmt.Errorf("BrasilAPI: request falhou: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- result{err: fmt.Errorf("BrasilAPI: status inesperado: %s", resp.Status)}
		return
	}

	var data struct {
		CEP          string `json:"cep"`
		State        string `json:"state"`
		City         string `json:"city"`
		Neighborhood string `json:"neighborhood"`
		Street       string `json:"street"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- result{err: fmt.Errorf("BrasilAPI: decode falhou: %w", err)}
		return
	}

	ch <- result{
		addr: Address{
			CEP:          data.CEP,
			State:        data.State,
			City:         data.City,
			Neighborhood: data.Neighborhood,
			Street:       data.Street,
			Source:       "BrasilAPI",
		},
	}
}

func fetchViaCEP(ctx context.Context, cep string, ch chan<- result) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		ch <- result{err: fmt.Errorf("ViaCEP: falha ao criar request: %w", err)}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- result{err: fmt.Errorf("ViaCEP: request falhou: %w", err)}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- result{err: fmt.Errorf("ViaCEP: status inesperado: %s", resp.Status)}
		return
	}

	var data struct {
		CEP        string `json:"cep"`
		UF         string `json:"uf"`
		Localidade string `json:"localidade"`
		Bairro     string `json:"bairro"`
		Logradouro string `json:"logradouro"`
		Erro       bool   `json:"erro"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		ch <- result{err: fmt.Errorf("ViaCEP: decode falhou: %w", err)}
		return
	}

	if data.Erro {
		ch <- result{err: errors.New("ViaCEP: CEP não encontrado")}
		return
	}

	ch <- result{
		addr: Address{
			CEP:          data.CEP,
			State:        data.UF,
			City:         data.Localidade,
			Neighborhood: data.Bairro,
			Street:       data.Logradouro,
			Source:       "ViaCEP",
		},
	}
}
