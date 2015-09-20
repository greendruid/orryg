package main

import "net/http"

func writeJSON(w http.ResponseWriter, data []byte) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, err := w.Write(data)

	return err
}

func handleCopierList(w http.ResponseWriter, req *http.Request) error {
	return writeJSON(w, []byte(`[]`))
}
