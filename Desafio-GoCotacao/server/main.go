// server.go
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiURL         = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	listenAddr     = ":8080"
	sqliteDSN      = "file:quotes.db?_pragma=busy_timeout(5000)"
	createTableSQL = `
CREATE TABLE IF NOT EXISTS quotes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bid TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);`
)

type awesomeResp struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type bidResponse struct {
	Bid string `json:"bid"`
}

// initDB() — Inicialização do SQLite
func initDB() (*sql.DB, error) {
	// sql.Open("sqlite", sqliteDSN) abre o pool de conexões.
	db, err := sql.Open("sqlite", sqliteDSN)
	if err != nil {
		return nil, err
	}
	// db.Exec(createTableSQL) garante a existência de quotes.
	if _, err := db.Exec(createTableSQL); err != nil {
		_ = db.Close()
		return nil, err
	}
	/*
		Retorna *sql.DB.
			Ponto importante: *sql.DB é thread-safe e gerencia pool internamente →
			pode ser reutilizado por múltiplos handlers concorrentes.
	*/
	return db, nil
}

// fetchUSDBRL(ctx) — Chamada externa com timeout 200ms
func fetchUSDBRL(ctx context.Context) (string, error) {
	// Cria um child context: ctxFetch, cancel := context.WithTimeout(ctx, 200*time.Millisecond).
	ctxFetch, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	// Monta http.NewRequestWithContext(ctxFetch, GET, apiURL, nil).
	req, err := http.NewRequestWithContext(ctxFetch, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}

	// http.Client sem timeout global (usamos o do context)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("[fetch] timeout (200ms) ao chamar AwesomeAPI: %v", err)
		}
		return "", err
	}
	defer resp.Body.Close()

	// Checa StatusCode (≥400 → erro).
	if resp.StatusCode >= 400 {
		log.Printf("[fetch] status inesperado da AwesomeAPI: %d", resp.StatusCode)
		return "", errors.New("awesomeapi returned non-2xx")
	}

	// json.NewDecoder(resp.Body).Decode(&ar) e valida ar.USDBRL.Bid.
	var ar awesomeResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", err
	}
	if ar.USDBRL.Bid == "" {
		return "", errors.New("campo bid vazio na resposta")
	}
	return ar.USDBRL.Bid, nil
	/*
		Efeitos de timeout:
			Se a AwesomeAPI demorar >200ms, Do() retorna erro com context.DeadlineExceeded.
			O handler detecta isso e propaga status 504 (detalhe abaixo).
			Racional do 200ms: requisito do desafio. É curto de propósito; você precisa
			lidar com timeouts de rede “do lado do servidor”.
	*/
}

// persistQuote(ctx, db, bid) — INSERT com timeout 10ms (fail) ou 50ms (success)
func persistQuote(ctx context.Context, db *sql.DB, bid string) error {
	// Cria child context: ctxDB, cancel := context.WithTimeout(ctx, 10*time.Millisecond).
	ctxDB, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	// Executa db.ExecContext(ctxDB, "INSERT INTO quotes (bid) VALUES (?)", bid).
	_, err := db.ExecContext(ctxDB, "INSERT INTO quotes (bid) VALUES (?)", bid)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("[db] timeout (10ms) ao persistir cotação: %v", err)
		}
		return err
	}
	/*
		Se o driver ficar “ocupado” (I/O, locks de página, etc.) e não concluir em 10ms:
			ExecContext retorna erro com context.DeadlineExceeded.
			Logamos: "[db] timeout (10ms)...".
			Mesmo com busy_timeout(5000) no DSN, o context vence: o ExecContext cancela
			a chamada antes dos 5s do SQLite.
			Política de resposta: persistência é “best-effort”. Se falhar, o servidor
			ainda responde o bid ao cliente
	*/
	return nil
}

// cotacaoHandler(db) — Orquestração da requisição
func cotacaoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Roteamento/validação: só aceita GET /cotacao, caso contrário 404.
		if r.Method != http.MethodGet || r.URL.Path != "/cotacao" {
			http.NotFound(w, r)
			return
		}

		/*
			Fetch externo (200ms): chama fetchUSDBRL(r.Context()).
				Falhou por timeout (DeadlineExceeded) → responde 504 Gateway Timeout.
				Falhou por outro motivo (status 5xx da AwesomeAPI, erro de parse etc.)
				 → responde 502 Bad Gateway.
		*/
		bid, err := fetchUSDBRL(r.Context())
		if err != nil {
			status := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) {
				status = http.StatusGatewayTimeout
			}
			http.Error(w, http.StatusText(status), status)
			return
		}

		/*
			Persistência (10ms): chama persistQuote(r.Context(), db, bid).
			Se falhar: loga e segue (não derruba a resposta).
		*/
		// Persistência best-effort: loga erro, mas ainda responde ao cliente
		if err := persistQuote(r.Context(), db, bid); err != nil {
			log.Printf("[db] erro ao persistir: %v", err)
		}

		/*
			Resposta:
				Content-Type: application/json.
				Body: {"bid":"<valor>"}.
				Status: 200 OK.
		*/
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(bidResponse{Bid: bid})
	}
}

// main() — Boot do servidor HTTP
func main() {
	// db := initDB().
	db, err := initDB()
	if err != nil {
		log.Fatalf("erro ao inicializar SQLite: %v", err)
	}
	defer db.Close()

	// mux := http.NewServeMux(); mux.Handle("/cotacao", cotacaoHandler(db)).
	mux := http.NewServeMux()
	mux.Handle("/cotacao", cotacaoHandler(db))

	/*
		Cria http.Server com timeouts de camada de transporte:
			ReadTimeout: 2s (limita leitura de request/headers).
			WriteTimeout: 2s (limita escrita da resposta).
			IdleTimeout: 30s (keep-alive).
	*/
	srv := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("server ouvindo em %s", listenAddr)
	// srv.ListenAndServe() inicia em :8080.
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("erro no servidor: %v", err)
	}
}

/*
	Por que também usar timeouts no http.Server?
		Eles protegem o servidor contra clientes lentos (slowloris) e travamentos na
		camada HTTP mesmo quando a lógica interna usa context.

	Tratamento de erros e códigos de status (decisão)
		504 Gateway Timeout: quando nós (o servidor) não conseguimos falar com o
		upstream (AwesomeAPI) no prazo (200ms). É o código semântico correto para “o meu upstream estourou tempo”.

		502 Bad Gateway: quando houve falha no upstream que não é timeout
		(p. ex., 500 da AwesomeAPI, erro de parse de JSON).

		200 OK: mesmo se o INSERT falhar — pois isso é “efeito colateral” opcional;
		o dado principal (bid) foi obtido com sucesso.

	Concurrency e segurança
		*sql.DB é seguro para concorrência.
		Cada request tem seu próprio context e seus deadlines.
		log.Printf é seguro multi-thread; mensagens podem intercalar, mas não corrompem.
		O arquivo quotes.db fica na pasta server/ (diretório de trabalho do processo).

	Pontos de atenção / edge cases
		Primeira execução do SQLite pode ser marginalmente mais lenta (alocação/DDL);
		10ms é intencionalmente apertado — por isso você viu timeouts em alguns
		cenários (comportamento esperado).

		Formato da AwesomeAPI: se mudasse o shape e USDBRL.bid ficasse ausente → devolvemos erro (502).

		Windows: caminho do DB depende do current working directory. Mantendo o go
		run dentro de server/, o arquivo fica lá.
*/
