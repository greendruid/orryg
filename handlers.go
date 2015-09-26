package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func writeJSON(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(data)

	return err
}

func handleCopierList(w http.ResponseWriter, req *http.Request) error {
	return writeJSON(w, []byte(`[]`))
}

func tmpl(name string) *template.Template {
	return template.Must(template.ParseFiles(fmt.Sprintf("./assets/templates/%s.html", name)))
}

func handleIndex(w http.ResponseWriter, req *http.Request) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	return tmpl("index").Execute(w, struct {
		Title string
	}{
		Title: "Accueil",
	})
}
