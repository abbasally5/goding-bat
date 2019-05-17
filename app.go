package main

import (
	"fmt"
	"html/template"
	"net/http"
)

// Handlers
var baseTemplates = []string{
	"templates/problem_sets.tmpl",
	"templates/layout/footer.tmpl",
	"templates/layout/header.tmpl",
	"templates/layout/base.tmpl",
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(baseTemplates...)

	if err != nil {
		fmt.Println(err)
		return
	}
	err = t.ExecuteTemplate(w, "problem_sets.tmpl", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", homeHandler)

	http.ListenAndServe(":8080", nil)
}
