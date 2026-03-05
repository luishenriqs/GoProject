# Desafio — Multithreading e APIs (Go Expert / Full Cycle)

Este repositório contém a solução do desafio do módulo **“Multithreading e APIs”**.

## Objetivo

Realizar **duas requisições simultâneas** (concorrentes) para buscar dados de um **CEP** em duas APIs diferentes e:

- **Aceitar a resposta mais rápida**.
- **Descartar a resposta mais lenta**.
- Exibir no **terminal** os dados do endereço **e qual API respondeu**.
- Aplicar **timeout de 1 segundo**. Se exceder, exibir erro de **timeout**.

APIs usadas:

- BrasilAPI: `https://brasilapi.com.br/api/cep/v1/<CEP>`
- ViaCEP: `http://viacep.com.br/ws/<CEP>/json/`

## Requisitos do desafio

- CEP fornecido diretamente via **argumento** (ex.: `01153000`)
- Sem template obrigatório
- Não exige testes automatizados

## Como executar

### 1) Pré-requisitos

- Go instalado (versão compatível com seu ambiente do curso)

### 2) Rodar o projeto

Na raiz do repositório:

```bash
go run . 01153000
