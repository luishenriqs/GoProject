# 9-deploy-k8s

## Objetivo do módulo

Este módulo introduz o uso de **Docker** e **Kubernetes** para executar uma aplicação Go em um ambiente orquestrado, com foco em:

- empacotar a aplicação em uma imagem Docker;
- criar um cluster Kubernetes local com `kind`;
- publicar a aplicação no cluster;
- expor a aplicação com `Service`;
- entender o papel de `Deployment`, `Pod`, `Service` e `replicas`;
- aplicar verificações de saúde com `startupProbe`, `readinessProbe` e `livenessProbe`;
- observar, na prática, o comportamento de **self-healing** do Kubernetes.

A aplicação usada no módulo é propositalmente simples:

```go
package main

import "net/http"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	http.ListenAndServe(":8080", nil)
}
```

Ela sobe um servidor HTTP na porta `8080` e responde `Hello World` na rota `/`.

---

## Visão geral do que foi aprendido

Ao longo deste módulo, foi possível observar os seguintes conceitos na prática:

1. **Docker em modo desenvolvimento**
   - uso de um container com Go instalado para rodar a aplicação localmente;
   - bind mount do diretório do projeto para dentro do container;
   - execução do `main.go` com `go run`.

2. **Docker em modo produção**
   - uso de build multi-stage para gerar um binário Go;
   - uso de imagem final mínima com `scratch`;
   - separação entre imagem de desenvolvimento e imagem de produção.

3. **Kubernetes local com kind**
   - criação de um cluster Kubernetes local usando Docker;
   - uso de `kubectl` para consultar o estado do cluster;
   - carregamento de imagem local no cluster com `kind load docker-image`.

4. **Deployment**
   - criação e manutenção automática de Pods;
   - escalabilidade via `replicas`;
   - substituição automática de Pods deletados manualmente.

5. **Service**
   - criação de um ponto de acesso estável para os Pods;
   - seleção dos Pods por `label`;
   - exposição da porta da aplicação.

6. **Health checks**
   - `startupProbe`: tempo de inicialização;
   - `readinessProbe`: se o container está pronto para receber tráfego;
   - `livenessProbe`: se o container continua saudável depois de iniciado.

7. **Troubleshooting**
   - erro de path incorreto (`kos/deployment.yaml`);
   - diferença entre aplicar `Deployment` e aplicar `Service`;
   - impacto de trocar a configuração do manifest;
   - erro `ErrImagePull` / `ImagePullBackOff` ao usar imagem local sem política adequada de pull;
   - limitação prática de `LoadBalancer` em ambiente `kind`.

---

## Estrutura dos arquivos do módulo

### `main.go`

Responsável por expor um servidor HTTP simples:

- escuta em `:8080`;
- responde `Hello World` em `/`.

### `Dockerfile`

Usado para desenvolvimento local:

```dockerfile
FROM golang:latest

WORKDIR /app

CMD ["go", "run", "main.go"]
```

Esse arquivo é útil para estudo e iteração rápida, porque:

- não gera binário;
- executa o código diretamente com `go run`;
- depende do código-fonte estar presente no container.

### `Dockerfile.prod`

Usado para build de produção:

```dockerfile
FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-w -s" -o server .

FROM scratch
COPY --from=builder /app/server .
CMD ["./server"]
```

Esse fluxo aplica **multi-stage build**:

#### Etapa 1 — builder
- usa uma imagem Go completa;
- copia o projeto;
- compila um binário Linux estático.

#### Etapa 2 — imagem final
- usa `scratch`, sem shell nem utilitários adicionais;
- copia apenas o binário final;
- reduz o tamanho da imagem e remove dependências desnecessárias.

Esse é o Dockerfile mais apropriado para uso com Kubernetes.

### `docker-compose.yaml`

Usado para rodar localmente com Docker Compose:

```yaml
services:
  goapp:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - .:/app
```

Esse arquivo foi útil na etapa inicial para:

- subir um container local;
- publicar a porta `8080`;
- montar o diretório do projeto dentro do container.

---

## O que é Kubernetes

Kubernetes é um orquestrador de containers.

Na prática, ele resolve problemas que surgem quando uma aplicação deixa de ser apenas “um container rodando localmente” e passa a exigir:

- múltiplas instâncias;
- reinício automático em caso de falha;
- exposição estável de serviços;
- atualização controlada;
- separação entre disponibilidade da aplicação e ciclo de vida individual dos containers.

Em vez de você gerenciar containers manualmente, o Kubernetes passa a gerenciar esse ambiente com base em **manifests declarativos**.

---

## O que é o kind

`kind` significa **Kubernetes IN Docker**.

Ele permite criar um cluster Kubernetes local usando containers Docker como nodes. Isso é muito útil para aprendizado, testes locais e experimentação sem necessidade de um cluster remoto.

### Criação do cluster

O comando utilizado foi:

```bash
kind create cluster --name=goexpert
```

### Resultado esperado

Após a criação do cluster:

- o contexto do `kubectl` passa a apontar para o cluster criado;
- um node local fica disponível;
- o comando abaixo deve retornar o node em estado `Ready`:

```bash
kubectl get nodes
```

---

## Fluxo completo utilizado no módulo

### 1. Validar ferramentas instaladas

```bash
docker --version
kubectl version --client
kind --version
```

### 2. Criar o cluster local

```bash
kind create cluster --name=goexpert
```

### 3. Confirmar o cluster

```bash
kubectl get nodes
kubectl config current-context
```

### 4. Gerar a imagem da aplicação

```bash
docker build -f Dockerfile.prod -t goapp:latest .
```

### 5. Carregar a imagem no cluster kind

```bash
kind load docker-image goapp:latest --name goexpert
```

Esse passo é importante porque a imagem local do Docker host **não fica automaticamente disponível** para o cluster `kind`.

### 6. Aplicar os manifests Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### 7. Validar recursos criados

```bash
kubectl get deployments
kubectl get pods
kubectl get services
```

### 8. Testar acesso à aplicação

Em um primeiro momento, foi usado `port-forward`:

```bash
kubectl port-forward service/goapp-service 8080:80
```

Depois, em outro terminal:

```bash
curl localhost:8080
```

Resultado esperado:

```text
Hello World
```

---

## Manifest final do Deployment

O `Deployment` evoluiu para este estado:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: goapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: goapp
  template:
    metadata:
      labels:
        app: goapp
    spec:
      containers:
        - name: goapp
          image: goapp:latest
          resources:
            limits:
              memory: "32Mi"
              cpu: "100m"

          startupProbe:
            httpGet:
              path: /
              port: 8080
            periodSeconds: 10
            failureThreshold: 10

          readinessProbe:
            httpGet:
              path: /
              port: 8080
            periodSeconds: 10
            failureThreshold: 2
            timeoutSeconds: 5

          livenessProbe:
            httpGet:
              path: /
              port: 8080
            periodSeconds: 10
            failureThreshold: 3
            timeoutSeconds: 5
            successThreshold: 1

          ports:
            - containerPort: 8080
```

### Explicação por bloco

#### `replicas: 3`
Instrui o Kubernetes a manter **3 Pods** da aplicação em execução.

Na prática, isso foi validado quando o comando:

```bash
kubectl get pods
```

mostrou 3 Pods `Running`.

#### `selector.matchLabels` e `template.metadata.labels`
Esses dois blocos conectam o `Deployment` aos Pods que ele gerencia.

O seletor:

```yaml
matchLabels:
  app: goapp
```

deve corresponder exatamente ao label dos Pods:

```yaml
labels:
  app: goapp
```

#### `image: goapp:latest`
Define a imagem do container.

Como a imagem foi criada localmente, o ambiente `kind` depende do comando:

```bash
kind load docker-image goapp:latest --name goexpert
```

#### `resources.limits`
Define limites máximos de uso de CPU e memória:

- memória: `32Mi`
- CPU: `100m`

Isso introduz o conceito de governança de recursos do container.

#### `ports`
Expõe a porta `8080` dentro do container.

---

## Probes: startup, readiness e liveness

Um dos conceitos mais importantes deste módulo foi a introdução das probes.

### 1. `startupProbe`

```yaml
startupProbe:
  httpGet:
    path: /
    port: 8080
  periodSeconds: 10
  failureThreshold: 10
```

Função:
- diz ao Kubernetes como verificar se a aplicação **conseguiu iniciar corretamente**;
- durante a fase de startup, falhas ainda são toleradas dentro do limite configurado.

No contexto da aplicação do módulo:
- a verificação é feita por HTTP em `/`;
- a porta verificada é a `8080`.

### 2. `readinessProbe`

```yaml
readinessProbe:
  httpGet:
    path: /
    port: 8080
  periodSeconds: 10
  failureThreshold: 2
  timeoutSeconds: 5
```

Função:
- informa se o container está **pronto para receber tráfego**;
- enquanto não estiver pronto, o `Service` não deve enviar requisições para ele.

Esse conceito é importante porque:
- um container pode estar em execução;
- mas ainda não estar pronto para atender de forma segura.

### 3. `livenessProbe`

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 8080
  periodSeconds: 10
  failureThreshold: 3
  timeoutSeconds: 5
  successThreshold: 1
```

Função:
- informa se o container **continua vivo e saudável** depois de iniciado;
- se o teste falhar repetidamente, o Kubernetes reinicia o container.

Isso é essencial para cenários em que a aplicação “trava” ou deixa de responder, mesmo sem encerrar o processo.

---

## Manifest final do Service

O `Service` evoluiu para este estado:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: goapp-service
spec:
  type: LoadBalancer
  selector:
    app: goapp
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

### Explicação por bloco

#### `selector`
Define que esse `Service` deve apontar para Pods com o label:

```yaml
app: goapp
```

#### `port`
É a porta do serviço Kubernetes.

#### `targetPort`
É a porta exposta pelo container dentro do Pod.

Nesse caso:
- o serviço recebe em `8080`;
- encaminha para o container em `8080`.

#### `type: LoadBalancer`
Esse tipo de `Service` é usado para expor a aplicação externamente em ambientes que possuam integração com um balanceador de carga real.

### Observação importante para ambiente local com kind

No `kind`, trocar o `Service` para `LoadBalancer` **não significa automaticamente** que você terá um IP externo funcional.

Em ambiente local com `kind`, o comportamento padrão normalmente não cria um balanceador real como ocorreria em um cloud provider. Portanto:

- o `Service` pode continuar sem `EXTERNAL-IP`;
- o acesso local ainda pode exigir `kubectl port-forward`;
- para laboratório local, `ClusterIP` e `port-forward` continuam sendo a forma mais previsível de teste.

---

## Conceitos observados na prática com os logs

### 1. Erro de path incorreto

Foi executado:

```bash
kubectl apply -f kos/deployment.yaml
```

e o retorno foi:

```text
error: the path "kos/deployment.yaml" does not exist
```

Aprendizado:
- o `kubectl apply` depende do path exato do arquivo;
- esse erro não indica problema no Kubernetes, mas apenas path incorreto no terminal.

### 2. Escalabilidade via `replicas`

Após aplicar o `Deployment` com `replicas: 3`, o retorno de `kubectl get pods` mostrou 3 Pods ativos.

Aprendizado:
- o `Deployment` mantém a quantidade desejada de réplicas;
- o Kubernetes agenda múltiplos Pods a partir do mesmo template.

### 3. Self-healing

Foi executado:

```bash
kubectl delete pod goapp-7757dbf9d8-7npw5
```

e logo depois um novo Pod foi criado.

Aprendizado:
- o Pod individual não é o recurso “desejado”;
- o `Deployment` é o recurso que define o estado desejado;
- se um Pod desaparecer, o Kubernetes cria outro automaticamente para restaurar o número de réplicas.

Esse é um dos comportamentos mais importantes do Kubernetes.

### 4. Diferença entre atualizar Deployment e Service

Aplicar novamente apenas o `deployment.yaml` não alterou o `Service`.

Aprendizado:
- `Deployment` e `Service` são recursos independentes;
- mudar um não altera automaticamente o outro;
- cada manifest precisa ser aplicado de forma explícita quando houver alteração.

### 5. `curl localhost:8080` retornando `Empty reply from server`

Esse comportamento apareceu após alteração do `Service`, mas sem `port-forward` ativo.

Aprendizado:
- criar ou alterar um `Service` não significa que a aplicação ficará automaticamente acessível no `localhost` do host;
- `localhost` da máquina do desenvolvedor não é o mesmo endpoint do cluster;
- em laboratório local com `kind`, o acesso via `curl localhost:8080` continua dependendo da forma de exposição escolhida.

### 6. `ErrImagePull` / `ImagePullBackOff`

Depois de reaplicar o `Deployment`, apareceu um novo Pod com:

- `ErrImagePull`
- `ImagePullBackOff`

enquanto os Pods antigos seguiam funcionando.

#### O que isso ensina

Quando o `Deployment` cria um novo Pod, o Kubernetes tenta obter a imagem definida no manifest.

Como a imagem é:

```yaml
image: goapp:latest
```

e o cluster está em ambiente local `kind`, podem ocorrer problemas se:

- a imagem não estiver carregada no cluster;
- a política de pull fizer o Kubernetes tentar buscar a imagem externamente;
- o `Pod` novo for criado com base em uma configuração que o cluster não consegue resolver.

#### Como evitar isso no kind

Em ambiente local com imagem carregada manualmente, o manifest deve normalmente incluir:

```yaml
imagePullPolicy: Never
```

ou, em alguns cenários, `IfNotPresent`.

No fluxo deste módulo, `imagePullPolicy: Never` é a opção mais segura para deixar explícito que a imagem local já foi carregada no cluster e não deve ser buscada em registry remoto.

---

## Diferença entre os principais recursos usados

### Pod
É a menor unidade executável do Kubernetes.  
No seu caso, cada Pod contém um container da aplicação Go.

### Deployment
É o recurso declarativo que gerencia Pods.

Responsabilidades:
- criar Pods;
- manter a quantidade de réplicas;
- substituir Pods removidos ou com falha;
- aplicar atualizações de template.

### Service
É o recurso que fornece um endpoint estável para acessar os Pods.

Responsabilidades:
- selecionar Pods por label;
- encaminhar tráfego para eles;
- abstrair o fato de que Pods podem morrer e renascer com nomes diferentes.

---

## Limitações e cuidados observados neste módulo

### 1. `LoadBalancer` em ambiente local
Em `kind`, o tipo `LoadBalancer` não se comporta como em um provedor cloud real. O uso didático é válido, mas o acesso local continua diferente do cenário gerenciado em nuvem.

### 2. Imagem local precisa estar disponível no cluster
Gerar a imagem com `docker build` não basta. Em `kind`, é necessário carregá-la explicitamente com:

```bash
kind load docker-image goapp:latest --name goexpert
```

### 3. `latest` pode dificultar diagnóstico
Usar `goapp:latest` é aceitável no laboratório, mas em fluxo real é melhor versionar imagens.

### 4. `scratch` reduz a imagem, mas dificulta depuração interna
A imagem final com `scratch` é ótima para produção, mas não contém shell ou ferramentas de inspeção. Isso é bom para enxugar a imagem, mas limita diagnósticos dentro do container.

---

## Fluxo recomendado para repetir o laboratório

### Subir/validar cluster

```bash
kind create cluster --name=goexpert
kubectl get nodes
```

### Buildar imagem

```bash
docker build -f Dockerfile.prod -t goapp:latest .
```

### Carregar imagem no kind

```bash
kind load docker-image goapp:latest --name goexpert
```

### Aplicar manifests

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

### Validar recursos

```bash
kubectl get deployments
kubectl get pods
kubectl get svc
```

### Testar localmente com port-forward

```bash
kubectl port-forward service/goapp-service 8080:8080
```

Em outro terminal:

```bash
curl localhost:8080
```

---

## Comandos úteis de diagnóstico

### Ver Pods

```bash
kubectl get pods
```

### Ver Services

```bash
kubectl get svc
```

### Descrever Pod

```bash
kubectl describe pod <nome-do-pod>
```

### Ver logs do container

```bash
kubectl logs <nome-do-pod>
```

### Deletar um Pod para observar self-healing

```bash
kubectl delete pod <nome-do-pod>
```

### Reaplicar manifests

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

---

## Melhorias naturais para a próxima etapa

Após este módulo, as próximas evoluções naturais seriam:

1. adicionar `imagePullPolicy: Never` ao `Deployment` para uso local com kind;
2. usar tags versionadas em vez de `latest`;
3. experimentar `NodePort` para exposição local sem depender de `port-forward`;
4. estudar `rolling update` ao alterar a imagem;
5. adicionar `requests` além de `limits`;
6. introduzir variáveis de ambiente e `ConfigMap`.

---

## Resumo executivo do módulo

Este módulo consolidou a transição de uma aplicação Go local para um cenário orquestrado com Kubernetes.

O principal aprendizado foi que:

- Docker empacota a aplicação;
- Kubernetes gerencia a execução declarativa dessa aplicação;
- `Deployment` define o estado desejado;
- `Service` fornece conectividade estável;
- probes ajudam o cluster a decidir quando iniciar, receber tráfego e reiniciar containers;
- o cluster recria Pods automaticamente;
- ambiente local com `kind` exige cuidados específicos com imagens locais e exposição de serviços.

Na prática, este módulo mostrou não apenas o “caminho feliz”, mas também erros reais de operação que ajudam a fixar o comportamento do Kubernetes: path incorreto, mudança parcial de manifest, conflito de acesso local e falha de pull de imagem.

---

## Estado final de referência

### `k8s/deployment.yaml`
- `Deployment`
- 3 réplicas
- container `goapp`
- imagem `goapp:latest`
- limites de CPU e memória
- `startupProbe`
- `readinessProbe`
- `livenessProbe`
- porta `8080`

### `k8s/service.yaml`
- `Service`
- nome `goapp-service`
- seletor `app: goapp`
- tipo `LoadBalancer`
- porta `8080` → `8080`

Esse conjunto representa o snapshot final do aprendizado deste módulo.
