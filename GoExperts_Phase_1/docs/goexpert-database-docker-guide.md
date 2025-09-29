# Guia Rápido: Banco de Dados com Docker + Integração com Go

Este documento orienta como configurar rapidamente um banco de dados (MySQL, PostgreSQL ou outro) utilizando Docker, acessá-lo via terminal, criar tabelas manualmente e estabelecer conexão via código Go.

---

## 📦 1. Subindo o Banco de Dados com Docker (MySQL)

### docker-compose.yaml

Crie um arquivo `docker-compose.yaml` no diretório do seu projeto com o seguinte conteúdo:

```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: goexpert
      MYSQL_USER: myuser
      MYSQL_PASSWORD: root
    ports:
      - "3306:3306"
    volumes:
      - mysql-data:/var/lib/mysql
    networks:
      - backend

volumes:
  mysql-data:

networks:
  backend:
```

### Comando para subir o container:

```bash
docker-compose up -d
```

Verifique se está rodando:

```bash
docker ps
```

---

## 🐚 2. Acessando o Banco via Terminal

Entre no container MySQL:

```bash
docker exec -it mysql bash
```

Conecte ao MySQL com as credenciais do compose:

```bash
mysql -u myuser -p
# Digite a senha: root
```

Selecione o banco:

```sql
USE goexpert;
```

Crie a tabela manualmente (exemplo):

```sql
CREATE TABLE products (
  id VARCHAR(255) PRIMARY KEY,
  name VARCHAR(255),
  price FLOAT
);
```

Verifique se foi criada:

```sql
SHOW TABLES;
```

---

## 🧠 3. Conectando ao Banco com Go

### Requisitos

```bash
go mod init seu_modulo
go get github.com/go-sql-driver/mysql
go get github.com/google/uuid
```

### main.go

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type Product struct {
	ID    string
	Name  string
	Price float64
}

func main() {
	dsn := "myuser:root@tcp(localhost:3306)/goexpert"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}
	defer db.Close()

	p := NewProduct("Notebook", 1999.90)
	if err := insertProduct(db, *p); err != nil {
		log.Fatalf("Erro ao inserir produto: %v", err)
	}
}

func NewProduct(name string, price float64) *Product {
	return &Product{
		ID:    uuid.New().String(),
		Name:  name,
		Price: price,
	}
}

func insertProduct(db *sql.DB, product Product) error {
	stmt, err := db.Prepare("INSERT INTO products (id, name, price) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(product.ID, product.Name, product.Price)
	return err
}
```

### Executar:

```bash
go run main.go
```

---

## 🛠️ 4. Verificando os dados inseridos

Reabra o terminal dentro do container ou use uma ferramenta como **Beekeeper** ou **DBeaver** com os seguintes dados:

- Host: `localhost`
- Porta: `3306`
- Usuário: `myuser`
- Senha: `root`
- Banco: `goexpert`

---

## 🔁 5. Comandos Úteis

```bash
docker-compose down         # Derruba os containers
docker-compose up -d       # Sobe os containers
docker exec -it mysql bash # Acessa o container
```

---

## 🧩 Observações

- O driver `mysql` no Go precisa estar importado anonimamente (`_ "github.com/go-sql-driver/mysql"`).
- Use `go mod tidy` sempre que adicionar dependências.
- O ID do produto é gerado com `uuid.New().String()`.

---

## 📚 Referências

- Curso GoExpert
- Docker Docs
- MySQL Docs
- Go Docs
