package main

import (
	"net/http"
	"strings"
	"text/template"
)

type Curso struct {
	Nome         string
	CargaHoraria int
}

type Cursos []Curso

func ToUpper(s string) string {
    return strings.ToUpper(s)
}

func main() {
    templates := []string{
        "header.html",
        "content.html",
        "footer.html",
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        t := template.New("content.html")
        t.Funcs(template.FuncMap{"ToUpper": ToUpper}) // Mapa de funções a serem parseadas
        t = template.Must(t.ParseFiles(templates...))
        err := t.Execute(w, Cursos{
            {"Go", 40},
            {"Java", 40},
            {"Python", 40},
            {"C#", 30},
        })
        if err != nil {
            panic(err)
        }
    })
    http.ListenAndServe(":8282", nil)
}

