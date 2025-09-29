# GoExpert — Phase 1 Consolidated Study Notes

> Este documento reúne todos os tópicos estudados na primeira fase do curso.

# 01 — File I/O Basics (os, bufio)

## Key Concepts
- `os.Create` creates/opens for writing (truncate if file exists). Always `Close()` or `defer file.Close()`.
- Writing:
  - `file.Write([]byte("..."))` for raw bytes.
  - `file.WriteString("...")` for strings.
- Reading:
  - `os.ReadFile(path)` → returns `[]byte`; simple for small files.
  - Streamed read: `os.Open` + `bufio.NewReader` + `Read(buffer)` for chunked reads.
- Deleting: `os.Remove(path)`.

## Cheat‑Sheet
```go
f, err := os.Create("Arquivo.txt"); if err != nil { panic(err) }
defer f.Close()
_, _ = f.Write([]byte("line 1\n"))
n, _ := f.WriteString("line 2")
fmt.Printf("written: %d bytes\n", n)

data, _ := os.ReadFile("Arquivo.txt")
fmt.Println(string(data))

rf, _ := os.Open("Arquivo.txt")
defer rf.Close()
r := bufio.NewReader(rf)
buf := make([]byte, 8)
for {
    n, err := r.Read(buf)
    if n > 0 { fmt.Print(string(buf[:n])) }
    if err != nil { break }
}

_ = os.Remove("Arquivo.txt")
```

## Pitfalls & Tips
- Always `defer Close()` as soon as you open a file.
- Check `n>0` before slicing buffer to avoid panics.
- Prefer `os.ReadFile` for small files and quick scripts; use buffered reads for large files.

## Exercises
- Write a function that copies a file using a 4KB buffer.
- Stream a large log file and count lines.

---

# 02 — HTTP Client (net/http)

## Key Concepts
- Simple fetch: `http.Get(url)` returns `*http.Response` (remember to `defer resp.Body.Close()`).
- Use `http.Client` to configure timeouts.
- Build requests with `http.NewRequest(method, url, body)` and customize headers.

## Cheat‑Sheet
```go
// 1) Quick GET
resp, err := http.Get("https://google.com")
if err != nil { panic(err) }
defer resp.Body.Close()
b, _ := io.ReadAll(resp.Body)
println(string(b))

// 2) Client with timeout
client := http.Client{ Timeout: time.Second }
resp, err = client.Get("https://google.com")
// ...

// 3) Custom request + headers
req, _ := http.NewRequest("GET", "https://google.com", nil)
req.Header.Set("Accept", "application/json")
resp, err = client.Do(req)
defer resp.Body.Close()
```

## Pitfalls & Tips
- **Always** close `resp.Body` to avoid leaks.
- Set a timeout on the client to prevent hanging calls.
- For JSON APIs, set `Content-Type: application/json` when sending bodies.

## Exercises
- Make a GET to a public JSON API and pretty‑print the result.
- Add retry logic with exponential backoff.

---

# 03 — JSON Encoding/Decoding (encoding/json)

## Key Concepts
- `json.Marshal` encodes a struct into `[]byte`.
- `json.Unmarshal` decodes `[]byte` into a struct (pass pointer).
- Tags control field names and exposure:
  - `json:"n"` rename field.
  - `json:"-"` omit field during encoding/decoding.

## Example Struct
```go
type Conta struct {
    Numero int `json:"n"`
    Saldo  int `json:"-"`  // omitted from JSON
}
```

## Cheat‑Sheet
```go
// Encode
c := Conta{Numero: 1, Saldo: 100}
b, _ := json.Marshal(c)
fmt.Println(string(b)) // {"n":1}

_ = json.NewEncoder(os.Stdout).Encode(c) // stream to stdout

// Decode
raw := []byte(`{"n":2,"s":200}`)
var out Conta
_ = json.Unmarshal(raw, &out)
fmt.Println(out.Numero, out.Saldo) // 2 0 (Saldo omitted)
```

## Pitfalls & Tips
- Only exported fields (capitalized) are encoded/decoded.
- When decoding unknown payloads, consider `json.Decoder` and `DisallowUnknownFields()` in strict scenarios.
- For large responses, prefer streaming with `Decoder` over loading all bytes.

## Exercises
- Implement strict decoding that fails on unknown fields.
- Write a helper to pretty‑print JSON with `json.MarshalIndent`.

---

# 04 — CEP Mini‑Project (HTTP + JSON + Files)

## Overview
Query ViaCEP, parse JSON into a struct, and write a short summary to a file.

## Data Model
```go
type ViaCEP struct {
    Cep, Logradouro, Complemento, Unidade, Bairro, Localidade, Uf, Estado,
    Regiao, Ibge, Gia, Ddd, Siafi string `json:"..."`
}
```

## CLI Version
```go
for _, cep := range os.Args[1:] {
    resp, _ := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)

    var data ViaCEP
    _ = json.Unmarshal(body, &data)

    f, _ := os.Create("Endereco.txt")
    defer f.Close()
    _, _ = f.WriteString(fmt.Sprintf("Localidade: %s", data.Localidade))
}
```

## HTTP Handler Version
```go
func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" { w.WriteHeader(http.StatusNotFound); return }
    cep := r.URL.Query().Get("cep")
    if cep == "" { w.WriteHeader(http.StatusBadRequest); return }

    data, err := getCep.GetCepFunc(cep)
    if err != nil { w.WriteHeader(http.StatusInternalServerError); return }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(data)
}
```

## getCep Package
```go
func GetCepFunc(cep string) (*ViaCEP, error) {
    resp, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
    if err != nil { return nil, err }
    defer resp.Body.Close()
    b, err := io.ReadAll(resp.Body)
    if err != nil { return nil, err }
    var out ViaCEP
    if err := json.Unmarshal(b, &out); err != nil { return nil, err }
    return &out, nil
}
```

## Pitfalls & Tips
- Validate `cep` format (length/digits) before calling API.
- Handle rate‑limits/timeouts; set `http.Client{Timeout: ...}`.
- Consider caching successful lookups to reduce API calls.

## Exercise
- Extend the handler to log requests and cache results in memory with TTL.

---

# 05 — HTTP Server, ServeMux & Static Files

## Multiple ServeMux/Ports
```go
muxA := http.NewServeMux()
muxB := http.NewServeMux()

muxA.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){ w.Write([]byte("Scheduler")) })
muxB.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){ w.Write([]byte("Appointment")) })

go http.ListenAndServe(":8080", muxA)
http.ListenAndServe(":8081", muxB)
```

## Static File Server
```go
fs := http.FileServer(http.Dir("./public"))
mux := http.NewServeMux()
mux.Handle("/", fs)
mux.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request){
    w.Write([]byte("Hello from blog"))
})
log.Fatal(http.ListenAndServe(":8080", mux))
```

## Minimal HTML (public/index.html)
```html
<!doctype html><html lang="en"><meta charset="utf-8"><title>Demo</title>
<h1>Olá Mundo!</h1>
```

## Pitfalls & Tips
- Prefer explicit `http.NewServeMux()` over default mux in bigger apps.
- For production, add timeouts: `Server{ReadTimeout, WriteTimeout, IdleTimeout}`.
- Validate paths to avoid directory traversal when serving files.

## Exercise
- Add graceful shutdown with `context` and `Server.Shutdown`.

---

# 06 — Database with `database/sql` (MySQL)

## Setup
- DSN: `<user>:<password>@tcp(<host>:<port>)/<dbname>`
- Use `_ "github.com/go-sql-driver/mysql"` for side‑effect import.

## Model
```go
type Product struct {
    ID    string
    Name  string
    Price float64
}
```

## CRUD Highlights
```go
// INSERT
stmt, _ := db.Prepare("INSERT INTO products (id, name, price) VALUES (?, ?, ?)")
defer stmt.Close()
_, _ = stmt.Exec(p.ID, p.Name, p.Price)

// UPDATE
stmt, _ = db.Prepare("UPDATE products SET name=?, price=? WHERE id=?")
_, _ = stmt.Exec(p.Name, p.Price, p.ID)

// SELECT one
stmt, _ = db.Prepare("SELECT id, name, price FROM products WHERE id=?")
var out Product
err := stmt.QueryRow(id).Scan(&out.ID, &out.Name, &out.Price)
if err == sql.ErrNoRows { /* handle not found */ }

// SELECT all
rows, _ := db.Query("SELECT id, name, price FROM products")
defer rows.Close()
for rows.Next() { /* rows.Scan(...); append */ }

// DELETE
stmt, _ = db.Prepare("DELETE FROM products WHERE id=?")
res, _ := stmt.Exec(id)
n, _ := res.RowsAffected()
```

## Pitfalls & Tips
- Always `defer rows.Close()` and `stmt.Close()`.
- Check `RowsAffected` for DELETE/UPDATE feedback.
- Use prepared statements for dynamic inputs; they help avoid SQL injection.
- Manage connections with `db.SetMaxOpenConns/SetMaxIdleConns/SetConnMaxLifetime` in real apps.

## Exercise
- Add transaction support to perform insert + update atomically.

---

# 07 — GORM ORM Essentials

## Models & Relations
```go
type Category struct {
    ID   uint   `gorm:"primaryKey"`
    Name string
}

type SerialNumber struct {
    ID        string `gorm:"type:char(36);primaryKey"`
    Number    string
    ProductID uint
}

type Product struct {
    gorm.Model
    Name         string  `gorm:"type:varchar(100);not null"`
    Price        float64 `gorm:"type:decimal(10,2);not null"`
    CategoryID   uint
    Category     Category
    SerialNumber SerialNumber `gorm:"constraint:OnDelete:CASCADE;"`
}
```

## Migrations
```go
if err := db.AutoMigrate(&Category{}, &Product{}, &SerialNumber{}); err != nil { /* ... */ }
```

## CRUD
```go
// Create
cat := &Category{Name: "Ferramentas"}
db.Create(cat)

prod := &Product{Name: "Machado", Price: 2989.90, CategoryID: cat.ID}
db.Create(prod)

sn := &SerialNumber{ID: uuid.New().String(), Number: "123456", ProductID: prod.ID}
db.Create(sn)

// Read (relations)
var products []Product
db.Preload("Category").Preload("SerialNumber").Find(&products)

// Update
prod.Price = 2999.00
db.Save(prod)

// Delete (with feedback)
res := db.Delete(&Product{}, "id = ?", prod.ID)
if res.RowsAffected == 0 { log.Println("no rows") }
```

## Pitfalls & Tips
- Put foreign‑key dependencies *after* parents in `AutoMigrate` order.
- `gorm.Model` includes `ID uint` by default; if you add custom `ID`, avoid conflicts.
- Use `Preload` for eager loading; beware of N+1 queries.
- For UUID PKs, define `type:char(36);primaryKey` and set value before `Create` (or use hooks).

## Exercise
- Add `unique index` to `SerialNumber.Number` and handle duplicate errors.

---

