# Go Cotação — Desafio Técnico GO_EXPERTS FULL CYCLE

Este projeto implementa dois programas em Go (`server` e `client`) que se comunicam via HTTP para obter a cotação do dólar, persistir no banco de dados SQLite e salvar o valor em arquivo de texto.

---

## Estrutura do Projeto

```

go-cotacao/
├─ client/
│  └─ main.go        # Cliente HTTP
├─ server/
│  ├─ main.go        # Servidor HTTP
│  └─ quotes.db      # Banco SQLite (criado automaticamente)
├─ go.mod
└─ .gitignore

````

---

## Requisitos

- Go 1.22+
- SQLite CLI (opcional, para inspecionar o banco):  
  - Linux: `sudo apt-get install sqlite3`  
  - macOS: `brew install sqlite`  
  - Windows: [Download oficial](https://www.sqlite.org/download.html)

---

## Execução

### 1. Rodar o servidor
Na pasta `server/`:

```bash
cd server
go run main.go
````

O servidor iniciará em `http://localhost:8080` com o endpoint `/cotacao`.

---

### 2. Rodar o cliente

Em outro terminal, na pasta `client/`:

```bash
cd client
go run main.go
```

O cliente fará uma requisição ao servidor, obterá a cotação atual (`bid`) e salvará no arquivo:

```
client/cotacao.txt
```

Formato do arquivo:

```
Dólar: 5.35
```

---

### 3. Verificar persistência no banco

Na pasta `server/`:

```bash
sqlite3 quotes.db ".tables"
# deve listar: quotes

sqlite3 quotes.db "SELECT id, bid, created_at FROM quotes;"
```

Exemplo de saída:

```
1|5.3502|2025-09-12 17:31:47
2|5.35  |2025-09-12 17:38:34
3|5.3484|2025-09-12 17:41:52
```

---

## Regras de Timeout

* **Fetch da API externa (AwesomeAPI):** 200ms
* **Persistência em SQLite:** 10ms (ajustável)
* **Requisição do cliente ao servidor:** 300ms

Se algum limite for excedido, o erro será logado mas o fluxo principal continua.

---

## Observações

* O banco `quotes.db` é criado automaticamente na pasta `server/`.
* Caso queira testar diferentes cenários, altere o valor de timeout no código ou use variável de ambiente (se configurado).
* Logs informam quando ocorre timeout ou erro em qualquer etapa.

```

---


## ✍️ Author
**Luís Henrique Pereira**  
Fullstack Developer | Prompt Engineering  
[GitHub](https://github.com/luishenriqs) | [LinkedIn](https://www.linkedin.com/in/luis-pereira-nodejs-react-native/)  