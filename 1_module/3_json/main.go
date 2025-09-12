package main

import (
	"encoding/json"
	"os"
)

type Conta struct {
    Numero int `json:"n"` // Tag do Go (Tipo alias)
    Saldo int `json:"-"` // Tag de traço omite valor
}

func main() {
    // TRANSFORMAR STRUCT EM JSON
    conta := Conta{Numero: 1, Saldo: 100}
    res, err := json.Marshal(conta) // json.Marshal transforma struct em json
    if err != nil {
        println(err)
    }
    println(string(res)) // Imprime o json criado pelo Marshal

    err = json.NewEncoder(os.Stdout).Encode(conta) // json.NewEncoder transforma struct em json / os.Stdout imprime
    if err != nil {
        println(err)
    }

    // TRANSFORMAR JSON EM STRUCT
    jsonPuro := []byte(`{"n": 2, "s": 200}`) // Cria um json a partir de um slice de bytes
    var contaX Conta // Cria uma instância de conta
    err = json.Unmarshal(jsonPuro, &contaX) // json.Unmarshal retorna json para struct Params: 1°: json, 2°: endereço memória
    if err != nil {
        println(err)
    }
    println(contaX.Numero)
    println(contaX.Saldo)
}
