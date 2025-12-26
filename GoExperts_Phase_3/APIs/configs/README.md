# Configuração da Aplicação — `configs/`

A pasta `configs/` concentra tudo o que é necessário para **carregar e organizar as configurações de execução** da API.

Nesta API, configurações como **banco de dados**, **porta do servidor** e **parâmetros de JWT** são tratadas como parte do *runtime* e ficam fora das camadas de domínio e de casos de uso.

---

## Objetivo desta camada

- **Carregar configurações** a partir de arquivo `.env`/variáveis de ambiente usando **Viper**.
- Expor uma estrutura (ex.: `Config`) com os parâmetros necessários para inicializar o sistema.
- Centralizar a configuração de autenticação **JWT** (incluindo a preparação do `JWTAuth` com `jwtauth/v5`).

> Regra do projeto (importante): o runtime deve importar e usar **apenas** `configs.LoadConfig` do pacote `configs/` para carregar configuração.

---

## O que você encontra aqui

De forma geral, `configs/` é a “fonte única” das configurações que o `cmd/server/main.go` precisa para montar a aplicação:

- parâmetros de banco (driver e credenciais);
- porta do servidor HTTP;
- segredo e expiração do JWT;
- instância/configuração do `JWTAuth` usada pelo middleware/autenticação.

---

## Variáveis de ambiente suportadas (exemplo de `.env`)

Exemplo de `.env` conforme padrão do projeto:

```env
DB_DRIVER=mysql
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=root
DB_NAME=apiGoDB

WEB_SERVER_PORT=8000

JWT_SECRET=secret
JWT_EXPIRESIN=300
```

### Significado dos principais parâmetros

- `DB_DRIVER`  
  Define qual driver de banco será usado (ex.: `mysql` ou `sqlite`).

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`  
  Credenciais e dados de conexão (usados quando o driver é MySQL).

- `WEB_SERVER_PORT`  
  Porta em que o servidor HTTP será iniciado.

- `JWT_SECRET`  
  Segredo usado para assinar e validar tokens JWT.

- `JWT_EXPIRESIN`  
  Tempo de expiração do token (em segundos).

---

## Como essa pasta se encaixa no fluxo da API

1. `cmd/server/main.go` chama `configs.LoadConfig(".")` para carregar as configurações.
2. Com a config carregada:
   - a conexão com o banco é inicializada (via `infra/database`, respeitando `DB_DRIVER`);
   - a configuração de JWT é usada para validar/gerar tokens (middleware/login);
   - o servidor HTTP é iniciado na porta configurada (`WEB_SERVER_PORT`).

Em resumo: `configs/` fornece o “*estado de execução*” necessário para que as demais camadas funcionem sem precisar conhecer detalhes de `.env` ou Viper.

---

## Relação com as demais camadas

- **Consumida por**: `cmd/server/` (montagem da aplicação).
- **Impacta**:
  - `infra/database/` (conexão conforme `DB_DRIVER`);
  - `api/middleware/` (validação JWT, extração de `sub`, expiração);
  - `api/handlers/` (login/geração de token, quando aplicável).
- **Não depende de**: domínio (`internal/entity`) e usecases (`internal/usecase`).
