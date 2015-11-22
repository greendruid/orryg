package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"net/http"
	_ "net/http/pprof"

	"github.com/vrischmann/userdir"
)

func getLogFile() io.Writer {
	dir := filepath.Join(userdir.GetDataHome(), "orryg")

	{
		fi, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(dir, 0700)
		} else if err != nil && !os.IsNotExist(err) {
			log.Printf("unable to create log directory %s. err=%v", dir, err)
			return ioutil.Discard
		} else {
			if !fi.IsDir() {
				log.Printf("unable to create log directory %s because it's already a file", dir)
				return ioutil.Discard
			}
		}
	}

	file := filepath.Join(dir, "main.log")
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Printf("unable to create log file %s. err=%v", file, err)
		return ioutil.Discard
	}

	return f
}

var (
	conf configuration
)

func main() {
	go http.ListenAndServe(":6060", nil)

	logger = log.New(io.MultiWriter(getLogFile(), os.Stdout), "orryg: ", log.LstdFlags)

	{
		conf = newWindowsConfiguration()

		s, err := conf.DumpConfig()
		if err != nil {
			logger.Printf("there was a problem while dumping the configuration. err=%v", err)
			return
		}

		logger.Printf("configuration dump")
		for _, line := range s {
			logger.Printf("%s", line)
		}
	}

	e := newEngine(conf)
	go e.run()

	err := buildUI()
	if err != nil {
		logger.Printf("unable to build the UI. err=%v", err)
		return
	}

	// NOTE(vincent): this is blocking
	mw.Run()

	if err = e.stop(); err != nil {
		logger.Printf("unable to stop engine correctly. err=%v", err)
	}
}
