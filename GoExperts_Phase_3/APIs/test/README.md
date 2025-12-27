# APIs/test — Artefatos de teste

Esta pasta centraliza **todos os artefatos de teste** do projeto (automatizados e manuais), mantendo-os separados do runtime da API.

## Conteúdo

### 1) Teste automatizado (E2E)

- `e2e_test.go`  
  Teste end-to-end/integrado usando `httptest.NewServer`, com foco em validar o contrato HTTP e o fluxo de autenticação/rotas protegidas.

### 2) Testes manuais no VS Code (arquivos `.http`)

Requisitos:
- VS Code com a extensão **REST Client** (botão **Send Request** aparece acima de cada request).

Arquivos:
- `user.http`  
  Coleção de requisições para o fluxo de **User**:
  - `POST /users`
  - `POST /login`
  - `GET /users/me`
  - `PUT /users/me`
  - `DELETE /users/me`

- `products.http`  
  Coleção de requisições para o fluxo de **Product**:
  - `POST /products`
  - `GET /products` (com paginação)
  - `GET /products/{id}`
  - `PUT /products/{id}`
  - `DELETE /products/{id}`

Uso recomendado:
1. Execute o `POST /login` para obter um token JWT.
2. Execute as demais rotas protegidas usando o token retornado.

> Observação: os arquivos `.http` foram estruturados para facilitar regressões manuais diretamente pelo editor.

### 3) Testes manuais no Bruno (coleção versionada)

- `bruno-collection-go-api/`  
  Coleção do **Bruno** versionada no repositório para executar e validar manualmente todos os endpoints (User e Product), incluindo rotas públicas e protegidas por JWT.

O diretório contém os artefatos padrão do Bruno, como:
- `bruno.json`
- `collection.bru`
- `environments/` (ex.: environment local apontando para `http://localhost:8000`)
- pastas por domínio (ex.: `User/`, `Product/`)

Como abrir no Bruno:
1. No Bruno, use **Open Collection** (ou equivalente).
2. Selecione a pasta `APIs/test/bruno-collection-go-api/`.
3. Selecione o environment (ex.: `GoLocal`) antes de executar as requests.

## Convenções

- Tudo aqui é **suporte de testes**: não deve ser importado pelo código de produção.
- Novos artefatos de teste (automatizados/manuais) devem ser adicionados nesta pasta para manter o padrão de organização do projeto.
