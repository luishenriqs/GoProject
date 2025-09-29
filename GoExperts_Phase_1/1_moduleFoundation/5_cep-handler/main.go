package main

import (
	"GoProject/1_moduleFoundation/5_cep-handler/getCep"
	"encoding/json"
	"net/http"
)


func main() {
	http.HandleFunc("/", BuscaCepHandler)
	http.ListenAndServe(":8080", nil) // Sobe servidor http
}

func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cepParam := r.URL.Query().Get("cep")
	if cepParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cep, error := getCep.GetCepFunc(cepParam) // usa a função modularizada
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Forma comentada: Se quiser posse do valor salvo em uma var use o .Marshal
	// result, err := json.Marshal(cep)
	// if err != nil {
	//     w.WriteHeader(http.StatusInternalServerError)
	//     return
	// }
	// w.Write(result)

	// Forma recomendada (direta):
	json.NewEncoder(w).Encode(cep)

}

// Criei a função o package getCep com a função GetCepFunc e modularizei o arquivo

// Para teste use o ThunderClient
// Para teste use o ThunderClient
// Para teste use o ThunderClient
// Para teste use o ThunderClient