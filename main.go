package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/vrischmann/flagutil"
)

var (
	flListenAddr flagutil.NetworkAddresses

	store *dataStore
	e     *engine
)

func init() {
	flag.Var(&flListenAddr, "l", "The listen address for the HTTP server")
}

func main() {
	flag.Parse()
	if len(flListenAddr) == 0 {
		flListenAddr = flagutil.NetworkAddresses{":8080"}
	}

	{
		var err error

		store, err = newDataStore()
		if err != nil {
			log.Fatalln(err)
		}

		err = store.init()
		if err != nil {
			log.Fatalln(err)
		}

		e, err = newEngine(store)
		if err != nil {
			log.Fatalln(err)
		}
		go e.run()
	}

	{
		router := newRouter()
		router.handleFunc("/copier/list", handleCopierList)
		go http.ListenAndServe(flListenAddr[0], router)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Kill, os.Interrupt)

	select {
	case <-signalCh:
	}

	log.Printf("stopping...")

	if err := e.stop(); err != nil {
		log.Fatalf("unable to stop engine. err=%v", err)
	}
	if err := store.Close(); err != nil {
		log.Fatalf("unable to close data store. err=%v", err)
	}
}
