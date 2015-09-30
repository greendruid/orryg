package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/codegangsta/negroni"
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
		router := http.NewServeMux()
		router.HandleFunc("/settings/list", handler(handleSettingsList))
		router.HandleFunc("/settings/change", handler(handleSettingsChange))
		router.HandleFunc("/copiers/list", handler(handleCopiersList))
		router.HandleFunc("/copiers/add", handler(handleCopiersAdd))
		router.HandleFunc("/copiers/remove/", handler(handleCopiersRemove))
		router.HandleFunc("/directories/list", handler(handleDirectoriesList))
		router.HandleFunc("/directories/add", handler(handleDirectoriesAdd))
		router.HandleFunc("/directories/remove/", handler(handleDirectoriesRemove))

		n := negroni.New()
		n.Use(negroni.NewLogger())
		n.UseHandler(router)

		go http.ListenAndServe(flListenAddr[0], n)
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
