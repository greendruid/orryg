package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type router struct {
	routes map[string]handlerFunc
}

func newRouter() *router {
	return &router{
		routes: make(map[string]handlerFunc),
	}
}

var (
	errUnauthorized = errors.New("unauthorized")
)

type handlerFunc func(w http.ResponseWriter, req *http.Request) error

func (r *router) handleFunc(path string, fn handlerFunc) {
	r.routes[path] = fn
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	lw := &loggingWriter{u: w}

	fn, ok := r.routes[req.URL.Path]
	if !ok {
		lw.WriteHeader(http.StatusNotFound)
		io.WriteString(lw, "Not Found")
	} else {
		err := fn(lw, req)
		if err == errUnauthorized {
			lw.WriteHeader(http.StatusUnauthorized)
			io.WriteString(lw, "Unauthorized")
		}
		if err != nil {
			lw.WriteHeader(http.StatusInternalServerError)
			io.WriteString(lw, "Internal Server Error")
			log.Printf("error while processing request. err=%v", err)
		}
	}

	elapsed := time.Now().Sub(start)

	log.Printf("%s %d %s - %s", elapsed, lw.st, http.StatusText(lw.st), req.URL.Path)
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
