# Go — Módulos Privados, Autenticação (GitHub/Bitbucket), Go Proxy e `vendor`

Este README registra os conceitos e o “como fazer” para:

- consumir **módulos privados**;
- configurar **GOPRIVATE** (e variáveis relacionadas);
- garantir **autenticação** (GitHub e Bitbucket) durante `go get` / `go mod tidy`;
- entender **Go Proxy** e **checksum database (sumdb)**;
- usar **`vendor`** quando fizer sentido.

> Exemplos usam comandos do Go 1.18+ (módulos como padrão). Os mesmos conceitos valem para versões mais novas.

---

## 1) Contexto rápido: como o Go resolve dependências

No Go, dependências são módulos versionados identificados por um **module path**, por exemplo:

- `github.com/minha-org/meu-modulo`
- `bitbucket.org/minha-org/meu-modulo`

Quando você executa comandos como:

- `go mod tidy`
- `go get <mod>@<versao>`
- `go test ./...`

...o Go precisa baixar versões dos módulos referenciados no `go.mod` (diretamente ou transitivamente).

O download acontece via:

1. **Go Proxy** (padrão: `proxy.golang.org`) — um cache/proxy de módulos.
2. Se o proxy não deve/ não pode ser usado (caso de privados), o Go busca **direto no VCS** (Git, etc.).
3. A verificação de integridade normalmente passa pela **sumdb** (padrão: `sum.golang.org`) — base de checksums.

Para módulos privados, você normalmente **desativa** proxy e sumdb para os domínios/paths privados. É aí que entra o `GOPRIVATE`.

---

## 2) Módulos privados: objetivo e por que exigem configuração

### O que é um módulo privado

É um módulo cujo código está em um repositório (GitHub/Bitbucket/etc.) que **exige credenciais** para leitura.

### Problema que aparece sem configuração

Sem ajustes, o Go tenta usar o proxy público (`proxy.golang.org`) e a sumdb pública (`sum.golang.org`).

- Para módulos privados, isso falha por **permissão**.
- Além disso, você geralmente **não quer** que um módulo privado seja consultado por serviços públicos, mesmo que apenas para resolução/verificação.

### Solução: marcar paths como privados

Você marca quais módulos são “privados” por padrão usando **`GOPRIVATE`** (e, consequentemente, o Go passa a buscar direto no VCS e não usa sumdb para esses paths).

---

## 3) `GOPRIVATE` (e variáveis relacionadas)

### 3.1) `GOPRIVATE`

`GOPRIVATE` define padrões (globs) de módulos que devem ser tratados como privados.

Exemplos típicos:

- Tudo dentro de uma organização no GitHub:
  ```bash
  go env -w GOPRIVATE=github.com/minha-org/*
  ```

- Tudo em um domínio interno:
  ```bash
  go env -w GOPRIVATE=git.minhaempresa.com/*
  ```

- Múltiplos padrões (separados por vírgula):
  ```bash
  go env -w GOPRIVATE=github.com/minha-org/*,bitbucket.org/minha-org/*
  ```

> Dica: use `go env GOPRIVATE` para conferir.

### 3.2) `GONOPROXY` e `GONOSUMDB`

O Go também expõe controles mais finos:

- `GONOPROXY`: padrões que **não** devem usar proxy.
- `GONOSUMDB`: padrões que **não** devem usar a sumdb.

Na prática, quando você define `GOPRIVATE`, o Go passa a tratar aqueles módulos como:

- sem proxy (equivalente a configurar `GONOPROXY` para os mesmos patterns);
- sem sumdb (equivalente a configurar `GONOSUMDB` para os mesmos patterns).

Você só costuma mexer diretamente em `GONOPROXY`/`GONOSUMDB` quando quer exceções específicas.

### 3.3) `GOPROXY` e `GOSUMDB`

- `GOPROXY` controla de onde o Go baixa módulos.
  - padrão comum: `https://proxy.golang.org,direct`
  - significa: tenta proxy público; se não achar, tenta direto.

- `GOSUMDB` controla a checksum database.
  - padrão comum: `sum.golang.org`

Para ambientes corporativos, é comum:

- trocar `GOPROXY` por um **proxy interno** (Artifactory, Athens, etc.);
- manter `direct` como fallback, dependendo da política;
- ajustar `GOPRIVATE` para garantir que módulos privados não vazem para proxy/sumdb públicos.

---

## 4) Autenticação: como o `go` acessa repositórios privados

Quando o Go busca um módulo privado diretamente, ele depende do **VCS** subjacente (normalmente Git). Ou seja:

- se `git clone` daquele repo funciona no seu terminal/ambiente, o Go tende a funcionar também;
- se `git clone` falha por credencial, o `go` também falha.

A seguir estão abordagens práticas para GitHub e Bitbucket.

---

## 5) GitHub: padrões de autenticação (recomendados)

### 5.1) Opção A — SSH (com chave)

Se você já usa SSH:

1. garanta que seu repo remoto use `git@github.com:org/repo.git`;
2. garanta que `ssh -T git@github.com` autentica (chave carregada).

O Go consegue baixar via Git usando SSH, desde que o módulo esteja referenciado de forma compatível.

> Observação: o module path no `go.mod` costuma ser `github.com/org/repo` (HTTP), mas o Git pode ser instruído a reescrever para SSH via config.

### 5.2) Opção B — HTTPS + token (PAT)

Para GitHub, senhas comuns deixaram de funcionar; o padrão é usar **Personal Access Token (PAT)**.

Você pode armazenar credenciais de forma segura usando o credential helper do Git.

Exemplo (varia por SO):

- Windows: Git Credential Manager
- macOS: Keychain
- Linux: libsecret / cache

Uma alternativa simples (menos segura) é usar `credential.helper store` (grava em texto). Em ambiente corporativo ou máquina compartilhada, evite.

### 5.3) Reescrever URLs (HTTP → SSH)

Uma técnica útil para forçar SSH mesmo que o módulo apareça como `https://github.com/...`:

```bash
git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
```

Com isso, quando o Go invocar o Git, o Git reescreve para SSH.

---

## 6) Bitbucket: padrões de autenticação

No Bitbucket, cenários comuns:

### 6.1) SSH

- Configure chave SSH e associe no Bitbucket.
- Garanta que `ssh -T git@bitbucket.org` funciona.

E, se necessário, use reescrita:

```bash
git config --global url."ssh://git@bitbucket.org/".insteadOf "https://bitbucket.org/"
```

### 6.2) HTTPS + App Password

Bitbucket costuma usar **App Password** para acesso via HTTPS.

O fluxo é análogo ao GitHub com PAT:

- usar o credential helper do Git;
- garantir que `git clone https://bitbucket.org/<org>/<repo>.git` funciona sem pedir senha a cada comando.

---

## 7) Checklist prático: “módulo privado funcionando”

1. Configure `GOPRIVATE` para os seus padrões privados:
   ```bash
   go env -w GOPRIVATE=github.com/minha-org/*,bitbucket.org/minha-org/*
   ```

2. Garanta que o Git está autenticado no mesmo ambiente onde você roda `go`:

   - `git clone` de um repo privado deve funcionar.

3. Rode comandos do Go:

   ```bash
   go clean -modcache
   go mod tidy
   go test ./...
   ```

4. Se falhar, rode com mais detalhe:

   ```bash
   GIT_TRACE=1 GIT_CURL_VERBOSE=1 go mod download
   ```

Isso ajuda a ver se o problema é:

- DNS / rede;
- proxy corporativo;
- credencial ausente;
- URL incorreta;
- `GOPRIVATE` não cobrindo o pattern certo.

---

## 8) Go Proxy: o que é e como entra no fluxo

### 8.1) O que o Go Proxy faz

Um Go Proxy (como o `proxy.golang.org`) é um serviço que:

- baixa módulos públicos uma vez;
- serve como cache global;
- melhora performance e disponibilidade;
- ajuda a tornar builds mais reprodutíveis (módulos versionados ficam “congelados” no proxy).

### 8.2) Por que existe o `,direct`

O valor típico `GOPROXY=https://proxy.golang.org,direct` significa:

- tente o proxy primeiro;
- se o módulo não existir lá, baixe direto do VCS.

Para módulos privados, você normalmente não quer nem tentar o proxy público. `GOPRIVATE` resolve isso.

### 8.3) Proxy corporativo (quando existe)

Empresas costumam colocar um proxy interno para:

- cache;
- auditoria/compliance;
- permitir builds em redes restritas;
- controlar dependências aceitas.

Nesse cenário, você pode ter algo como:

```bash
go env -w GOPROXY=https://meu-proxy-interno,direct
```

E ainda assim manter `GOPRIVATE` para que módulos privados sejam resolvidos de modo apropriado (ou mesmo também via proxy interno, dependendo da política).

---

## 9) `vendor`: o que é, por que é pouco usado, e quando faz sentido

### 9.1) O que é `vendor`

`vendor` é um diretório dentro do seu projeto contendo uma cópia das dependências.

Você gera esse diretório com:

```bash
go mod vendor
```

Isso cria:

- `vendor/` com os pacotes necessários;
- `vendor/modules.txt` descrevendo quais módulos foram vendorizados.

### 9.2) Como forçar build/test usando vendor

Você pode rodar:

```bash
go test -mod=vendor ./...
```

Ou configurar globalmente para o projeto/ambiente:

```bash
go env -w GOFLAGS=-mod=vendor
```

> Cuidado: `GOFLAGS=-mod=vendor` afeta todos os comandos Go naquele ambiente.

### 9.3) Por que `vendor` é pouco usado em muitos times

Porque o `go mod` já resolve muitos problemas de reprodutibilidade, e o Go Proxy/sumdb ajudam a manter módulos acessíveis. Além disso:

- aumenta o tamanho do repositório;
- gera churn de diffs (vendor muda quando dependências mudam);
- pode conflitar com políticas de “não commitar dependências”.

### 9.4) Quando `vendor` pode ser importante

Casos típicos:

- **Ambiente isolado/air-gapped** (sem acesso externo a VCS/proxies);
- **Compliance**: exigência de auditar e “congelar” dependências no repo;
- **Build hermético**: reduzir dependência de serviços externos no pipeline;
- **Incidentes de disponibilidade** (ex.: dependência sumiu/foi removida/indisponível) — `vendor` garante que o build ainda consegue compilar.

### 9.5) Fluxo recomendado ao usar `vendor`

1. Atualize dependências via `go get` / `go mod tidy`.
2. Gere `vendor`:
   ```bash
   go mod vendor
   ```
3. Rode validação usando vendor:
   ```bash
   go test -mod=vendor ./...
   ```
4. Faça commit do `go.mod`, `go.sum` e `vendor/` (se a política do projeto exigir).

---

## 10) Boas práticas de operação

- **Padronize `GOPRIVATE`** para sua org/domínio.
- **Use credential helper** (evite token em texto quando possível).
- **Valide com `go mod tidy` e `go test ./...`** após mudar credenciais/variáveis.
- Em ambiente corporativo, alinhe com o time:
  - se existe proxy interno (`GOPROXY` custom);
  - se `vendor` é política;
  - se há exigências de auditoria/supply chain.

---

## 11) Comandos úteis (resumo)

Ver ambiente Go:
```bash
go env
```

Configurar módulos privados:
```bash
go env -w GOPRIVATE=github.com/minha-org/*,bitbucket.org/minha-org/*
```

Limpar cache de módulos (para forçar re-download):
```bash
go clean -modcache
```

Baixar dependências:
```bash
go mod download
```

Organizar `go.mod`/`go.sum`:
```bash
go mod tidy
```

Gerar vendor:
```bash
go mod vendor
```

Testar usando vendor:
```bash
go test -mod=vendor ./...
```

---

## 12) Erros comuns e causas prováveis

- **401/403 ao baixar módulo privado**
  - Git não autenticado (SSH/PAT/App Password);
  - `GOPRIVATE` não cobre o path correto;
  - URL reescrita necessária (HTTP → SSH).

- **Tentativa de acessar proxy público para módulo privado**
  - `GOPRIVATE` ausente/mal configurado;
  - patterns incorretos (ex.: faltou `/*`).

- **Falhas em rede (proxy corporativo / SSL inspection)**
  - necessidade de configurar proxy do sistema;
  - certificados corporativos;
  - `GOPROXY` interno.

---

### Fim
