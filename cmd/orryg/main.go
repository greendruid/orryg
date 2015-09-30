package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
		Addr string `json:"addr"`
	}
	cl *client
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

	if !os.IsNotExist(err) {
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
	}

	if len(flAddr) > 0 {
		conf.Addr = flAddr[0]
	}

	return nil
}

func printUsage() {
	fmt.Println("Usage: orryg [options] <command> [arguments]\n")
	fmt.Println("  Available commands\n")
	fmt.Printf("%20s   %s\n", "copiers", "Manipulate the copiers")
	fmt.Printf("%20s   %s\n", "directories", "Manipulate the directories")
	fmt.Printf("%20s   %s\n", "settings", "Manipulate the settings")
	fmt.Println("\n\n  Global options")
	fmt.Printf("%20s   %s\n\n", "-h", "The adress of orrygd")
}

func main() {
	flag.Parse()

	err := parseConfigAndArgs()
	if err != nil {
		log.Fatalln(err)
	}

	if conf.Addr == "" {
		fmt.Println("please provide a server to connect to")
		os.Exit(1)
		return
	}

	if flag.NArg() < 1 {
		printUsage()
		os.Exit(1)
		return
	}

	cl, err = newClient(conf.Addr)
	if err != nil {
		log.Fatalln(err)
	}

	switch v := strings.ToLower(flag.Arg(0)); v {
	case "copiers":
		err = copiersCommand(flag.Args()[1:]...)
	case "directories":
		err = directoriesCommand(flag.Args()[1:]...)
	case "settings":
		err = settingsCommand(flag.Args()[1:]...)
	default:
		err = fmt.Errorf("unknown command '%s'\n", v)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
