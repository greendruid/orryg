package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		fmt.Println("Usage: bin2hex <package> [<variable name>=<file>] [<variable name 2>=<file 2>]")
		os.Exit(1)
	}

	pkg := flag.Arg(0)

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "package %s\n\n", pkg)

	for _, pair := range flag.Args()[1:] {
		tokens := strings.Split(pair, "=")
		if len(tokens) < 2 {
			fmt.Printf("Expected a pair of type <variable name>=<file name> but got '%s'", pair)
			os.Exit(1)
		}

		variable, filename := tokens[0], tokens[1]

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("unable to read file. err=%v", err)
		}

		fmt.Fprintf(&buf, `var %s = [...]byte{`, variable)

		for i, b := range data {
			fmt.Fprintf(&buf, "0x%02X", b)
			if i+1 < len(data) {
				buf.WriteString(", ")
			}
		}

		buf.WriteString("}\n\n")
	}

	o, err := os.Create("resources_generated.go")
	if err != nil {
		log.Fatalf("unable to create generated resources file. err=%v", err)
	}
	defer o.Close()

	io.Copy(o, &buf)
}
