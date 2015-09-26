package main

import (
	"fmt"
	"net/http"
)

func writeJSON(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(data)

	return err
}

func writeError(w http.ResponseWriter, format string, args ...interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := fmt.Fprintf(w, format, args...)

	return err
}

func handleCopiersList(w http.ResponseWriter, req *http.Request) error {
	return writeJSON(w, []byte(`[]`))
}

func handleCopiersAdd(w http.ResponseWriter, req *http.Request) error {
	return writeJSON(w, []byte(`[]`))
}

func handleDirectoriesList(w http.ResponseWriter, req *http.Request) error {
	return writeJSON(w, []byte(`[]`))
}
