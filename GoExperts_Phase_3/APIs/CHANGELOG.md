# CHANGELOG

Este arquivo registra a **evolução do projeto** ao longo do tempo, documentando **novos avanços, features e melhorias** adicionadas ao repositório.

## Objetivo

- Manter um histórico **claro e rastreável** do que foi alterado.
- Explicar **o que foi adicionado**, **como foi integrado** e **qual o impacto no comportamento**.
- Facilitar auditoria, revisão e retomada de contexto em novas sessões de trabalho.

## Convenções

- Formato inspirado em **Keep a Changelog**: cada versão possui seções **Added**, **Changed**, **Fixed**, **Removed** (conforme aplicável).
- Datas no formato `YYYY-MM-DD` (fuso Brasília).
- O README é a **documentação viva**; este arquivo é o **registro histórico**.

---

## [Unreleased]
### Added
- (a preencher)

### Changed
- (a preencher)

### Fixed
- (a preencher)

---

## [V6] - 2025-12-26

### Added
- Middleware de **logging HTTP** (observabilidade) integrado ao servidor via `chimw.Logger` (Chi middleware), mantendo roteamento em `http.ServeMux`.
- **Swagger (OpenAPI) + Swagger UI**:
  - UI em `GET /swagger/index.html`
  - JSON em `GET /swagger/doc.json`
  - Artefatos gerados em `docs/` via `go generate` (padronizado por `make swagger`):
    - `docs/docs.go`
    - `docs/swagger.json`
    - `docs/swagger.yaml`
- **Automação via Makefile** para execução e manutenção:
  - `make dev` (swagger → tests → run)
  - `make dev-fast` (tests → run)
  - `make run`, `make test`, `make swagger`, `make fmt`, `make clean-swagger`.
- **Testes manuais padronizados**:
  - Coleção Bruno em `test/bruno-collection-go-api/`
  - Requests `.http` em `test/user.http` e `test/product.http` (VS Code + REST Client).

### Changed
- Atualização do **composition root** (`cmd/server/main.go`) para incluir:
  - registro/serving das rotas do Swagger no `http.ServeMux`
  - wrapping do handler com middleware de logging antes do `ListenAndServe`.

### Notes
- Política adotada: a pasta `docs/` (Swagger) **é versionada** e deve ser mantida sincronizada via `make swagger`/`make dev` sempre que rotas/contratos mudarem.

---

## [V5] - 2025-12-24

### Fixed
- Correções editoriais internas no README (incluindo consistência de versão) e ajuste do diagrama Mermaid para renderização garantida.
