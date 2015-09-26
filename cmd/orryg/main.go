package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/vrischmann/flagutil"
	"github.com/vrischmann/userdir"
)

type command interface {
	Run(args ...string) error
}

type commandFunc func(args ...string) error

func (f commandFunc) Run(args ...string) error {
	return f(args...)
}

var (
	flAddr flagutil.NetworkAddresses

	conf struct {
		Addr string
	}
)

func init() {
	flag.Var(&flAddr, "h", "The address of orrygd")
}

func parseConfigAndArgs() error {
	configPath := filepath.Join(userdir.GetConfigHome(), "orryg", "config.json")
	f, err := os.Open(configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("%s is a directory", configPath)
	}

	dec := json.NewDecoder(f)
	if err := dec.Decode(&conf); err != nil {
		return err
	}

	if len(flAddr) > 0 {
		conf.Addr = flAddr[0]
	}

	return nil
}

func main() {
	flag.Parse()

	if err := parseConfigAndArgs(); err != nil {
		log.Fatalln(err)
	}
}
