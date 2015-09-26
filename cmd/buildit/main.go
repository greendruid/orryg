package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
)

const (
	jsDir = "./assets/js"

	devReact  = "react-dev.js"
	prodReact = "react.js"
)

var (
	flProd  bool
	flWatch bool
)

func init() {
	flag.BoolVar(&flProd, "prod", false, "Build for production")
	flag.BoolVar(&flWatch, "w", false, "Watch for changes")
}

func build() error {
	var buf bytes.Buffer

	files := []string{
		"ui.jsx",
		"app.jsx",
	}

	if flProd {
		fmt.Println("building with react prod")
		files = append([]string{prodReact}, files...)
	} else {
		fmt.Println("building with react dev")
		files = append([]string{devReact}, files...)
	}

	fmt.Printf("building %v\n", files)

	for _, el := range files {
		f, err := os.Open(filepath.Join(jsDir, el))
		if err != nil {
			return err
		}
		defer f.Close()

		io.Copy(&buf, f)
		buf.WriteByte('\n')
	}

	cmd := exec.Command("babel", "-o", "./assets/dist/app.js")
	cmd.Stdin = &buf
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func main() {
	flag.Parse()

	if flWatch {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatalln(err)
			return
		}
		defer w.Close()

		var wg sync.WaitGroup
		wg.Add(1)

		var timer *time.Timer

		go func() {
			for {
				select {
				case event := <-w.Events:
					if timer == nil {
						timer = time.AfterFunc(time.Millisecond*50, func() {
							if event.Op&fsnotify.Write == fsnotify.Write {
								if err := build(); err != nil {
									log.Fatalln(err)
								}
							}
							timer = nil
						})
					}
				case err := <-w.Errors:
					log.Fatalln(err)
				}
			}
			wg.Done()
		}()

		if err := w.Add(jsDir); err != nil {
			log.Fatalln(err)
		}

		wg.Wait()

		return
	}

	if err := build(); err != nil {
		log.Fatalln(err)
	}
}
