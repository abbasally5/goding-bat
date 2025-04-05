package main

import (
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	routes := http.NewServeMux()
	//routes.HandleFunc("/", homeHandler)
	//routes.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	//routes.HandleFunc("/prob/{probId:p[0-9]+}", probHandler)
	//routes.HandleFunc("/{setName:[a-zA-z]+\\-[0-9]+}", setHandler)

	http.ListenAndServe(":8080", routes)
}
