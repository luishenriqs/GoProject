# Swagger (OpenAPI) — Documentação da API

Este diretório (`/docs`) contém os **artefatos gerados automaticamente** pelo **Swaggo (swag)** para documentar a API via Swagger UI.

> **Importante:** os arquivos `docs.go`, `swagger.json` e `swagger.yaml` são **gerados**. Não edite manualmente esses arquivos — regenere sempre via `make swagger` / `go generate`.

---

## O que foi implementado

### 1) Geração automática do Swagger com `swag`
- A ferramenta **swag** foi instalada e usada para gerar a documentação a partir de anotações (`// @...`) no código.
- O projeto passou a usar **`go generate`** para rodar a geração do swagger de forma padronizada.

### 2) Integração do Swagger UI no servidor da API
- A API serve o Swagger UI em runtime, permitindo acessar a documentação em:
  - **Swagger UI:** `http://localhost:8000/swagger/index.html`
  - **Spec JSON:** `http://localhost:8000/swagger/doc.json`

> Obs.: para isso, o servidor importa o pacote gerado `docs` e registra o handler do swagger (ex.: `httpSwagger.WrapHandler`).

### 3) Automação via Makefile
Foi adicionado um `Makefile` na raiz do repositório com os comandos:

- `make swagger`  
  Gera os arquivos do swagger em `/docs` (sempre do zero, pois depende de `clean-swagger`).

- `make dev`  
  Fluxo completo: `swagger` → `test` → `run`.

- `make dev-fast`  
  Fluxo rápido: `test` → `run` (não regenera swagger).

- `make clean-swagger`  
  Remove os arquivos gerados do swagger.

- `make fmt`  
  Executa `gofmt -w .`.

---

## Como a geração funciona (go:generate)

O arquivo `cmd/server/main.go` contém um `//go:generate` que chama o `swag init` e **varre os diretórios corretos** (handlers, DTOs etc).

O ponto mais crítico é que o `swag` precisa “enxergar” os pacotes onde estão:
- handlers (rotas e anotações)
- DTOs/entidades usadas nas definições
- estruturas referenciadas pelo Swagger

Por isso, a lista de diretórios no `-d` aponta para pastas que **possuem arquivos `.go`**.

---

## Como usar no dia a dia

### Subir a API + Swagger (fluxo completo)
```bash
make dev
```

Depois acesse:
- `http://localhost:8000/swagger/index.html`

### Subir a API sem regenerar o Swagger (mais rápido)
```bash
make dev-fast
```

### Gerar/atualizar apenas os arquivos do Swagger
```bash
make swagger
```

### Apagar arquivos gerados do Swagger (regen limpo)
```bash
make clean-swagger
make swagger
```

---

## Quando o Swagger “fica vazio” (sem endpoints)

Se o Swagger UI abrir, mas aparecer **“No operations defined in spec!”** ou `doc.json` vier com `paths: {}`, quase sempre é porque:

1) O swagger foi gerado a partir do diretório errado (o `swag` não varreu os handlers), **ou**  
2) O swagger servido no servidor está desatualizado.

A correção padrão é:

```bash
make swagger
make dev
```

E validar:
- `http://localhost:8000/swagger/doc.json` contém `paths` preenchido.

---

## Observações sobre Windows / Git Bash

### Pastas estranhas `*;C`
Em alguns cenários no Windows, podem aparecer pastas como `swagger.json;C/` dentro de `docs/`.
Elas não fazem parte do projeto e podem ser removidas com:

```bash
rm -rf docs/*';C'/
```

### Warning do `mProfCycleWrap`
Pode aparecer um warning similar a:

```
warning: failed to evaluate const mProfCycleWrap ... reflect.Value.Len on zero Value
```

Isso é um ruído comum no Windows ao analisar constantes do runtime e **não impede** a geração do Swagger se os arquivos forem gerados corretamente.

---

## Artefatos gerados (não versionados manualmente)

Após `make swagger`, este diretório contém:

- `docs.go`  
  Código Go gerado que registra o `SwaggerInfo`.

- `swagger.json`  
  Especificação Swagger (JSON).

- `swagger.yaml`  
  Especificação Swagger (YAML).

---

## Referência rápida

- Regenerar swagger: `make swagger`
- Subir dev completo: `make dev`
- Abrir Swagger UI: `http://localhost:8000/swagger/index.html`
- Ver spec JSON: `http://localhost:8000/swagger/doc.json`
