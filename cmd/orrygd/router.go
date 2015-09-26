package main

import (
	"errors"
	"io"
	"log"
	"net/http"
)

var (
	errUnauthorized = errors.New("unauthorized")
)

type handlerFunc func(w http.ResponseWriter, req *http.Request) error

func handler(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// start := time.Now()

		err := fn(w, req)
		if err == errUnauthorized {
			w.WriteHeader(http.StatusUnauthorized)
			io.WriteString(w, "Unauthorized")
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Internal Server Error")
			log.Printf("error while processing request. err=%v", err)
		}

		// elapsed := time.Now().Sub(start)
	}
}

type loggingWriter struct {
	u  http.ResponseWriter
	st int
}

func (w *loggingWriter) Header() http.Header { return w.u.Header() }

func (w *loggingWriter) Write(p []byte) (n int, err error) {
	return w.u.Write(p)
}

func (w *loggingWriter) WriteHeader(st int) {
	w.st = st
	w.u.WriteHeader(st)
}
