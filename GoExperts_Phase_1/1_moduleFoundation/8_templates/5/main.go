package main

import (
	"net/http"
	"text/template"
)

type Curso struct {
	Nome         string
	CargaHoraria int
}

type Cursos []Curso

func main() {
    templates := []string{
        "header.html",
        "content.html",
        "footer.html",
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        t := template.Must(template.New("content.html").ParseFiles(templates...))
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

