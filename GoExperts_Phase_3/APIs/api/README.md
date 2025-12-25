# Camada Web/HTTP — `api/`

A pasta `api/` é a **porta de entrada HTTP** da API. Ela concentra tudo o que é específico de **HTTP**, como:
- roteamento de endpoints no `http.ServeMux`;
- leitura e escrita de **JSON**;
- autenticação via **JWT** (middleware);
- tradução de sucessos/erros em **status HTTP** e respostas JSON.

Nesta API (objetivo acadêmico), a camada Web é propositalmente “na unha”, usando apenas `net/http` (sem frameworks), para deixar explícito como roteamento, middleware e handlers se conectam no fluxo de uma requisição.

---

## Estrutura da camada

- `api/routes.go`  
  Registra os endpoints no `http.ServeMux` e **aplica o middleware** nas rotas protegidas.

- `api/handlers/`  
  Implementa os **handlers HTTP** (adaptação HTTP ↔ DTOs): decodifica request, chama usecases e formata a resposta.

- `api/middleware/`  
  Implementa o **middleware de autenticação JWT** e utilitários relacionados ao `context` (ex.: obter `userID` do contexto).

---

## Responsabilidades (o que cada parte faz)

### 1) `handlers/` — Adaptadores HTTP (JSON ↔ Usecases)

Handlers são responsáveis por transformar a requisição HTTP em uma chamada de caso de uso (usecase), e transformar o retorno em resposta HTTP.

Na prática, um handler faz:

1. **Ler** dados do request (ex.: `r.Body`, `r.URL`, headers).
2. **Decodificar JSON** (quando aplicável) e validar/transformar para DTOs (`internal/dto`).
3. **Chamar o usecase** correspondente (`internal/usecase`), passando apenas os inputs necessários.
4. **Traduzir o resultado** em:
   - status HTTP (ex.: `201`, `200`, `204`, `400`, `401`, `404`, `500`);
   - JSON de resposta (quando existir body).

> Observação importante do projeto: como o roteamento usa `http.ServeMux`, **o mux decide apenas pelo path** (não por método).  
> Em paths que aceitam múltiplos verbos (ex.: `/products`), o handler/wrapper precisa fazer o “dispatch por método” com `switch r.Method`.

---

### 2) `middleware/` — “Catraca” de autenticação (JWT)

O middleware JWT atua como uma barreira: **sem token válido, a requisição não chega aos handlers/usecases**.

O middleware faz:

1. Ler o header `Authorization`.
2. Extrair o token no formato `Bearer <token>`.
3. Validar assinatura **HS256** e validar expiração.
4. Extrair o `userID` do claim **`sub`**.
5. Colocar o `userID` no `context` da requisição e encaminhar para o próximo handler.

#### JWT: detalhes relevantes para o projeto
- Algoritmo: **HS256**
- Claims mínimos:
  - `sub`: ID do usuário autenticado (userID)
  - `exp`: expiração

#### Um detalhe importante para testes
O middleware injeta `userID` no `context` usando uma **chave privada do próprio package**.  
Isso impede “forjar” contexto com chaves inventadas em testes e, como consequência, força testes mais realistas:
- rotas protegidas devem ser testadas **passando pelo middleware real** e usando um token JWT real.

---

### 3) `routes.go` — Registro de rotas no `ServeMux` + aplicação de middleware

`api/routes.go` é o ponto central onde:
- os **paths** são registrados no `http.ServeMux`;
- rotas **públicas** são expostas diretamente;
- rotas **protegidas** são envolvidas pelo middleware JWT.

Isso deixa o “mapa de endpoints” visível e facilita entender, rapidamente:
- quais rotas exigem token;
- quais handlers estão por trás de cada endpoint.

---

## Contrato HTTP (endpoints)

A API expõe endpoints para duas entidades: `User` e `Product`.

### Rotas públicas (sem token)
- `POST /users`  
  Cria um usuário (cadastro).

- `POST /login`  
  Autentica email/senha e retorna `access_token` (JWT).

### Rotas protegidas (exigem JWT)
> Todas as rotas abaixo precisam do header:  
> `Authorization: Bearer <access_token>`

#### Endpoints “me” (usuário autenticado)
- `GET /users/me`  
  Retorna dados do usuário autenticado.

- `PUT /users/me`  
  Atualiza dados do usuário autenticado (campos opcionais).

- `DELETE /users/me`  
  Remove a conta do usuário autenticado.

#### Produtos
- `POST /products`  
  Cria um produto.

- `GET /products`  
  Lista produtos (paginado) com ordenação por `created_at`.

- `GET /products/{id}`  
  Busca produto por ID.

- `PUT /products/{id}`  
  Atualiza produto por ID (campos opcionais).

- `DELETE /products/{id}`  
  Remove produto por ID.

---

## Fluxo mental de uma requisição protegida (visão “de ponta a ponta”)

1. Cliente chama endpoint protegido enviando `Authorization: Bearer <token>`.
2. `ServeMux` resolve o handler pelo **path** (`routes.go`).
3. Middleware JWT valida token, extrai `sub`, injeta `userID` no `context`.
4. Handler:
   - lê `userID` do `context` (quando necessário);
   - decodifica JSON (quando aplicável);
   - chama o usecase;
   - traduz o resultado para status/JSON.
5. Usecase orquestra regras e chama o repositório (infra).
6. Repositório executa operações no banco via GORM.

Uma forma útil de revisar esse fluxo é lembrar das “funções” de cada etapa:
- **middleware** = barreira de autenticação
- **handler** = transformação HTTP ↔ DTO
- **usecase** = orquestração do caso de uso
- **repo** = persistência (objeto ↔ banco)

---

## Pontos de atenção (para evitar regressões)

- **ServeMux não roteia por método**  
  Se um path aceita múltiplos verbos (`GET/POST/PUT/DELETE`), o handler precisa fazer `switch r.Method`.

- **Padronização de JWT**  
  O projeto está padronizado em `jwtauth/v5`. Misturar versões (`jwtauth` sem `/v5` com `jwtauth/v5`) costuma causar incompatibilidade de tipos.

- **Context key privada no middleware**  
  Testes de rotas protegidas devem usar o middleware real e token real (evitar “forjar” contexto manualmente).

---

## Relação com as demais camadas

- `api/handlers` consome DTOs (`internal/dto`) e chama usecases (`internal/usecase`).
- `api/middleware` protege rotas e fornece `userID` via `context` para handlers.
- `api/routes.go` conecta paths → middleware → handlers.
- A composição (instanciação e wiring) acontece no **Composite Root**: `cmd/server/main.go`.
