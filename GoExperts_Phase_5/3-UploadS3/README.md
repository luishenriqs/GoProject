# README — Upload de arquivos para Amazon S3 com Go

## Visão geral do módulo

Este módulo de estudo teve como objetivo demonstrar um fluxo completo de **geração de arquivos locais** e **upload concorrente para um bucket Amazon S3 usando Go**.

A estrutura trabalhada foi enxuta e composta essencialmente por dois arquivos:

- `cmd/generator/main.go`
- `cmd/uploader/main.go`

Além deles, o projeto contém:
- pasta `tmp/`, usada para armazenar os arquivos gerados localmente
- `go.mod` e `go.sum`, responsáveis pelas dependências do módulo
- um arquivo `.csv` com credenciais exportadas, usado apenas como apoio local durante o estudo

---

## Estrutura do projeto

```text
3-UploadS3/
├── cmd/
│   ├── generator/
│   │   └── main.go
│   └── uploader/
│       └── main.go
├── tmp/
├── go.mod
└── go.sum
````

---

## Objetivo prático do capítulo

O capítulo ensinou como construir, em Go, um fluxo simples de processamento em lote com arquivos, cobrindo:

1. criação de múltiplos arquivos locais
2. leitura de arquivos de um diretório
3. integração com a AWS S3
4. upload de arquivos com a SDK oficial da AWS
5. uso de concorrência com goroutines
6. sincronização com `sync.WaitGroup`
7. limitação de paralelismo
8. retries em caso de falha
9. tratamento básico de erros
10. diferença entre código didático e código de produção

---

## Parte 1 — Geração de arquivos locais

O primeiro programa, localizado em `cmd/generator/main.go`, teve a responsabilidade de criar arquivos de teste dentro da pasta `./tmp`.

### Comportamento implementado

O programa:

* percorre um laço de `0` até `999`
* cria arquivos com nomes sequenciais, como:

  * `file0.txt`
  * `file1.txt`
  * `file2.txt`
  * ...
  * `file999.txt`
* escreve o conteúdo `"Hello, World!"` em cada arquivo
* fecha cada arquivo após a escrita

### Conceitos aprendidos nessa etapa

#### 1. Criação de arquivos com `os.Create`

Foi usado `os.Create` para criar ou sobrescrever arquivos:

```go
f, err := os.Create(fmt.Sprintf("./tmp/file%d.txt", i))
```

Esse trecho mostrou:

* como construir nomes dinamicamente com `fmt.Sprintf`
* como criar arquivos em disco a partir do Go
* como lidar com retorno de erro

#### 2. Escrita em arquivos com `WriteString`

Foi usado:

```go
_, err = f.WriteString("Hello, World!")
```

Com isso, aprendemos:

* como gravar conteúdo textual em um arquivo
* como verificar falha de escrita
* como descartar o valor retornado quando ele não é relevante

#### 3. Fechamento explícito do arquivo

O código fecha o arquivo com:

```go
err = f.Close()
```

Isso foi importante para reforçar:

* liberação correta do recurso
* persistência adequada do conteúdo em disco
* prevenção de vazamento de descritores de arquivo

#### 4. Fluxo síncrono simples

Essa parte do módulo usou um fluxo totalmente sequencial, ideal para iniciar o exercício:

* cria arquivo
* escreve
* fecha
* passa para o próximo

Essa simplicidade foi importante para preparar a pasta `tmp/` para a etapa seguinte.

---

## Parte 2 — Upload para Amazon S3

O segundo programa, localizado em `cmd/uploader/main.go`, foi responsável por ler os arquivos da pasta `tmp/` e enviar cada um para um bucket S3.

---

## Amazon S3: conceito abordado

Neste módulo, foi introduzido o Amazon S3 como serviço de armazenamento de objetos da AWS.

### O que foi entendido sobre o S3

* arquivos são armazenados como objetos
* cada objeto possui uma chave (`Key`)
* objetos são armazenados dentro de buckets
* o bucket funciona como um contêiner lógico
* para fazer upload, é necessário:

  * autenticar na AWS
  * definir a região
  * conhecer o bucket de destino
  * ter permissão de escrita

---

## Uso da AWS SDK for Go v2

Durante o estudo, a implementação foi adaptada para a **AWS SDK for Go v2**, que é a versão mais atual.

### Pacotes principais usados

* `github.com/aws/aws-sdk-go-v2/config`
* `github.com/aws/aws-sdk-go-v2/credentials`
* `github.com/aws/aws-sdk-go-v2/service/s3`

### Conceitos aprendidos

#### 1. Carregamento de configuração com `config.LoadDefaultConfig`

A configuração da AWS foi inicializada com:

```go
config.LoadDefaultConfig(...)
```

Isso mostrou como:

* carregar as configurações base da SDK
* definir a região
* injetar credenciais explicitamente quando necessário
* criar uma base para instanciar clientes dos serviços AWS

#### 2. Criação do cliente S3

Foi usado:

```go
s3.NewFromConfig(cfg)
```

Com isso aprendemos:

* como instanciar um client S3 na v2
* como reaproveitar a configuração carregada
* como separar a etapa de configuração da etapa de uso do serviço

#### 3. Bucket configurado em variável global

O bucket foi armazenado em uma variável global, permitindo:

* reutilização nas chamadas de upload
* centralização da configuração
* simplificação do código didático

---

## Leitura dos arquivos do diretório `tmp`

Antes de subir os arquivos, o programa precisou percorrer a pasta local.

### Abordagem utilizada

Foi feito:

```go
dir, err := os.Open("./tmp")
files, err := dir.Readdir(-1)
```

### Conceitos aprendidos

#### 1. Abrir diretório com `os.Open`

Em Go, diretórios também podem ser abertos como descritores, permitindo leitura posterior.

#### 2. Ler arquivos com `Readdir`

O uso de `Readdir(-1)` serviu para:

* obter todos os arquivos de uma vez
* iterar facilmente sobre o conteúdo
* ignorar subdiretórios quando necessário

#### 3. Filtragem de diretórios

Durante o loop, foi aplicado:

```go
if file.IsDir() {
    continue
}
```

Isso garantiu que apenas arquivos fossem enviados ao S3.

---

## Upload unitário com `PutObject`

O envio individual de cada arquivo foi feito com `PutObject`.

### Estrutura conceitual usada

* `Bucket`: nome do bucket
* `Key`: nome do arquivo no S3
* `Body`: conteúdo do arquivo aberto localmente

Exemplo conceitual:

```go
_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
    Bucket: &s3Bucket,
    Key:    &filename,
    Body:   f,
})
```

### Conceitos aprendidos nessa etapa

#### 1. O upload ocorre a partir de um `io.Reader`

O arquivo aberto com `os.Open` pode ser passado diretamente no campo `Body`.

#### 2. A chave do objeto pode ser o próprio nome do arquivo

Neste módulo, a `Key` escolhida foi o próprio nome do arquivo local, sem subpastas ou prefixos adicionais.

#### 3. O contexto controla a operação

Foi usado `context.WithTimeout(...)` para limitar o tempo da chamada, evitando que um upload fique bloqueado indefinidamente.

---

## Timeout com `context.WithTimeout`

Foi introduzido um timeout por upload.

### Objetivo

Evitar que uma chamada de rede fique pendurada sem fim.

### Conceito aprendido

```go
ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
defer cancel()
```

Isso ensinou:

* como criar um contexto com tempo limite
* por que é importante cancelar o contexto
* como aplicar timeout em chamadas externas

---

## Concorrência com goroutines

Depois do fluxo básico, o upload foi evoluído para execução concorrente.

### Conceito central

Cada upload passou a ser executado em uma goroutine.

Exemplo conceitual:

```go
go uploadFile(...)
```

### O que aprendemos

* goroutines permitem executar tarefas em paralelo de forma leve
* uploads independentes são um bom caso de uso para concorrência
* concorrência pode melhorar muito o throughput do processamento em lote

---

## Controle de concorrência

Não basta subir tudo em paralelo sem limite. Por isso, o estudo evoluiu para um controle de concorrência.

### Estratégia usada

Foi definido um número máximo de uploads simultâneos, por exemplo:

```go
const maxConcurrentUploads = 100
```

Depois, foi criado um mecanismo de pool de workers com um canal de tarefas.

### Resultado conceitual

* o sistema evita abrir uploads ilimitados ao mesmo tempo
* o consumo de recursos fica controlado
* o fluxo fica mais previsível

---

## Pool de workers

A implementação final do uploader adotou o padrão de worker pool.

### Como funciona

1. a `main` lê todos os arquivos
2. cada arquivo vira uma tarefa (`uploadTask`)
3. as tarefas são enviadas para um canal
4. vários workers ficam consumindo esse canal
5. cada worker executa `processUpload`

### Estrutura conceitual da tarefa

```go
type uploadTask struct {
    filename string
    attempt  int
}
```

Essa estrutura permitiu:

* transportar o nome do arquivo
* controlar o número de tentativas de envio

---

## Sincronização com `sync.WaitGroup`

Como várias goroutines são disparadas, foi necessário garantir que a aplicação só finalize quando tudo terminar.

### Conceito aprendido

* `wg.Add(1)` registra uma nova unidade de trabalho
* `wg.Done()` sinaliza que essa unidade terminou
* `wg.Wait()` bloqueia até todas as unidades terminarem

Isso foi essencial para evitar que a `main` termine antes dos uploads concluírem.

---

## Retry em caso de falha

O uploader foi evoluído para suportar novas tentativas quando ocorre erro.

### Estratégia usada

* cada tarefa possui um contador de tentativa
* quando o upload falha, a tarefa pode ser reenfileirada
* o retry só acontece até um limite máximo

Exemplo conceitual:

```go
if task.attempt < maxRetryAttempts {
    jobs <- uploadTask{
        filename: task.filename,
        attempt:  task.attempt + 1,
    }
}
```

### O que foi aprendido

* falhas transitórias podem ser tratadas com retry
* é necessário limitar a quantidade de tentativas
* retry sem limite pode gerar loop infinito

---

## Erros encontrados e aprendizados importantes

Durante o módulo, vários problemas práticos apareceram e foram úteis para consolidar os conceitos.

### 1. Escrita após `Close()`

Em uma versão inicial do gerador, a escrita estava sendo feita depois do `f.Close()`. Isso mostrou que:

* a ordem das operações importa
* o arquivo precisa estar aberto no momento da escrita

### 2. Mistura de abordagens de leitura de diretório

Foi experimentado:

* `Readdir(-1)` para ler tudo
* `Readdir(1)` dentro de um `for` para ler um por um

Misturar os dois no mesmo fluxo gerou inconsistência. O aprendizado foi:

* escolher uma única estratégia por implementação
* evitar reposicionar implicitamente o ponteiro de leitura do diretório sem necessidade

### 3. Problema de escopo com variável `files`

Uma versão intermediária tentou acessar `files[0]` fora do escopo correto. Isso reforçou:

* a importância do escopo de variáveis em Go
* a necessidade de manter a lógica linear e previsível

### 4. Retry com nome de arquivo incorreto

Em um momento, o retry reenfileava o caminho completo do arquivo em vez do nome simples. Isso fazia o código montar caminhos duplicados, como:

```text
./tmp/./tmp/file1.txt
```

Aprendizado:

* a estrutura da tarefa deve carregar exatamente o dado esperado
* path completo e nome do arquivo não são equivalentes

### 5. Loop infinito de retry

Também ficou claro que:

* erro permanente, como `AccessDenied`
* não deve ser tratado como erro transitório sem critério

Isso mostrou a importância de:

* limitar tentativas
* diferenciar tipos de erro
* evitar retries cegos

### 6. Erro `403 AccessDenied`

Ao executar com bucket/permissão inadequados, apareceu:

```text
StatusCode: 403
api error AccessDenied: Access Denied
```

Isso ensinou que:

* o código pode estar correto e ainda assim falhar por configuração externa
* permissões IAM e nome do bucket são parte fundamental do fluxo
* credenciais, região e policy do bucket precisam estar alinhadas

---

## Credenciais e configuração AWS

No estudo, chegou a ser usada configuração hardcoded para simplificar a execução local.

### Aprendizado importante

Embora funcione em ambiente didático, em projetos reais o ideal é:

* não deixar credenciais no código
* usar variáveis de ambiente
* usar profiles da AWS
* usar IAM Roles quando estiver em ambiente AWS

Também foi importante entender que:

* credencial válida sem permissão suficiente continua falhando
* bucket incorreto gera erro mesmo com código certo
* região incorreta pode quebrar a integração

---

## Diferença entre código didático e código de produção

Um ponto importante consolidado neste módulo foi a diferença entre um exemplo de aula e uma solução pronta para produção.

### No código didático

É aceitável simplificar:

* uso de variáveis globais
* credenciais temporariamente hardcoded
* logs simples com `fmt.Printf`
* estrutura concentrada em um único arquivo

### Em produção

O ideal seria evoluir para:

* injeção de dependência
* configuração externa
* logger estruturado
* tratamento de erro por tipo
* observabilidade
* testes automatizados
* retry inteligente
* separação em camadas e pacotes

---

## Fluxo completo aprendido no módulo

O fluxo final do módulo pode ser resumido assim:

### Etapa 1 — gerar massa de teste

O `generator` cria centenas de arquivos `.txt` na pasta `tmp`.

### Etapa 2 — inicializar AWS

O `uploader` carrega região, credenciais e bucket.

### Etapa 3 — ler a pasta local

Os arquivos da pasta `tmp` são listados.

### Etapa 4 — enfileirar tarefas

Cada arquivo vira uma tarefa de upload.

### Etapa 5 — processar em paralelo

Workers concorrentes consomem as tarefas.

### Etapa 6 — enviar ao S3

Cada worker abre o arquivo e executa `PutObject`.

### Etapa 7 — tratar falhas

Se houver erro, o arquivo pode ser reenfileirado até o limite de tentativas.

### Etapa 8 — aguardar conclusão

A `main` espera todas as tarefas terminarem com `WaitGroup`.

---

## Comandos típicos de execução

### Gerar arquivos locais

```bash
go run cmd/generator/main.go
```

### Executar uploader

```bash
go run cmd/uploader/main.go
```

---

## Pré-requisitos para o uploader funcionar

Para o uploader funcionar corretamente, é necessário:

* bucket S3 existente
* bucket com nome correto no código
* região correta
* credenciais válidas
* permissão `s3:PutObject`
* pasta `./tmp` contendo arquivos

---

## Conceitos de Go consolidados neste capítulo

Este módulo reforçou vários fundamentos importantes da linguagem:

### Entrada e saída de arquivos

* `os.Create`
* `os.Open`
* `WriteString`
* `Close`

### Formatação de strings

* `fmt.Sprintf`
* `fmt.Printf`

### Tratamento de erros

* verificação explícita com `if err != nil`
* propagação com `fmt.Errorf`

### Contexto

* `context.Background`
* `context.WithTimeout`

### Concorrência

* goroutines
* canais
* worker pool
* paralelismo controlado

### Sincronização

* `sync.WaitGroup`

### Estruturas

* definição de `struct` para representar tarefa

### Boas práticas básicas

* fechar arquivos
* controlar tentativas
* limitar simultaneidade
* separar responsabilidades

---

## O que este capítulo ensinou na prática

Ao final deste módulo, foi aprendido como:

* gerar arquivos automaticamente com Go
* percorrer arquivos de um diretório
* abrir arquivos para leitura
* enviar arquivos ao Amazon S3
* usar AWS SDK for Go v2
* criar uploads concorrentes
* controlar o número de tarefas simultâneas
* sincronizar goroutines com segurança
* aplicar timeout por operação
* implementar retry com limite
* diagnosticar erros reais de integração com a AWS

---

## Conclusão

Este capítulo foi importante porque conectou vários conhecimentos em um único exercício prático:

* manipulação de arquivos locais
* integração com serviço externo
* concorrência em Go
* controle de execução
* tratamento de falhas

Mesmo com poucos arquivos no módulo, o conteúdo foi rico em conceitos e mostrou um fluxo muito próximo de cenários reais de backend e processamento assíncrono.

O principal valor do capítulo foi demonstrar que Go permite construir pipelines de processamento simples, rápidos e claros, com excelente suporte nativo para concorrência e integração com serviços externos como o Amazon S3.

```