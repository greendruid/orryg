package main

import (
	"log"
	"os"
	"os/signal"
)

var (
	e    *engine
	conf config
)

func main() {
	var err error
	e, err = newEngine()
	if err != nil {
		log.Fatalln(err)
	}
	go e.run()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Kill, os.Interrupt)

	select {
	case <-signalCh:
	}

	log.Printf("stopping...")

	if err := e.stop(); err != nil {
		log.Fatalf("unable to stop engine. err=%v", err)
	}
}
