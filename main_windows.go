package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

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
	en   *engine // TODO(vincent): remove global maybe ?

	// standard flags
	flMinimized bool

	// debug flags
	flDebugResetLastUpdated bool
)

func init() {
	flag.BoolVar(&flMinimized, "minimized", false, "Start minimized")
	flag.BoolVar(&flDebugResetLastUpdated, "reset-last-updated", false, "Reset the last updated date of all directories")
}

func main() {
	flag.Parse()

	logger = log.New(io.MultiWriter(getLogFile(), os.Stdout), "orryg: ", log.LstdFlags)

	{
		conf = newWindowsConfiguration()

		if flDebugResetLastUpdated {
			dirs, err := conf.ReadDirectories()
			if err != nil {
				logger.Printf("unable to read the directories from the configuration. err=%v", err)
				return
			}

			for _, d := range dirs {
				d.LastUpdated = time.Time{}
				err := conf.UpdateLastUpdated(d)
				if err != nil {
					logger.Printf("unable to reset the last updated time. err=%v", err)
					return
				}
			}
		}

		s, err := conf.DumpConfig()
		if err != nil {
			logger.Printf("unable to dump the configuration. err=%v", err)
			return
		}

		logger.Printf("configuration dump")
		for _, line := range s {
			logger.Printf("%s", line)
		}
	}

	en = newEngine(conf)
	go en.run()

	err := buildUI()
	if err != nil {
		logger.Printf("unable to build the UI. err=%v", err)
		return
	}

	if flMinimized {
		mw.SetVisible(false)
	} else {
		mw.SetVisible(true)
	}
	// NOTE(vincent): this is blocking
	mw.Run()

	if err = en.stop(); err != nil {
		logger.Printf("unable to stop engine correctly. err=%v", err)
	}
}
