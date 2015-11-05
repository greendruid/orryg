package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/peterh/liner"
	"github.com/vrischmann/shlex"
)

var configurePromptCompletionNames = []string{
	"scp-copier", "directory", "check-frequency", "cleanup-frequency", "date-format",
}

type configurePrompt struct {
	conf configuration
	line *liner.State
}

func (p *configurePrompt) run() {
	p.line = liner.NewLiner()
	defer p.line.Close()

	p.line.SetCtrlCAborts(true)

	p.line.SetCompleter(func(line string) (c []string) {
		for _, n := range configurePromptCompletionNames {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	})

	for {
		l, err := p.line.Prompt("> ")
		if err == liner.ErrPromptAborted {
			return
		} else if err != nil {
			fmt.Printf("error while reading line. err=%v", err)
			return
		}

		tokens := shlex.Parse(l)
		if len(tokens) < 1 {
			fmt.Printf("no command given")
			continue
		}

		cmd := tokens[0]
		switch strings.ToLower(cmd) {
		case "scp-copier":
			p.runSCPCopierConfigureCommand(tokens[1:])
		case "directory":
			p.runDirectoryConfigureCommand(tokens[1:])
		case "check-frequency":
			p.runCheckFrequencyConfigureCommand(tokens[1:])
		case "cleanup-frequency":
			p.runCleanupFrequencyConfigureCommand(tokens[1:])
		case "date-format":
			p.runDateFormatConfigureCommand(tokens[1:])
		}
	}
}

func (p *configurePrompt) runSCPCopierConfigureCommand(args []string) {
	if len(args) == 0 {
		copiers, err := p.conf.ReadSCPCopiers()
		if err != nil {
			fmt.Printf("unable to read copiers. err=%v\n", err)
			return
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "copiers:\n")
		for _, copier := range copiers {
			fmt.Fprintf(&buf, "%s\n", copier)
		}

		fmt.Println(buf.String())

		return
	}

	if len(args) != 6 {
		fmt.Println("Usage: scp-copier <copier name> <user> <host> <port> <private key file> <backups dir>")
		return
	}

	name := args[0]
	user := args[1]
	host := args[2]
	sport := args[3]
	privateKeyFile := args[4]
	backupsDir := args[5]

	port, err := strconv.Atoi(sport)
	if err != nil {
		fmt.Printf("port '%s' is not valid\n", sport)
		return
	}

	err = p.conf.WriteSCPCopier(scpCopierConf{
		Name: name,
		Params: sshParameters{
			User:           user,
			Host:           host,
			Port:           int(port),
			PrivateKeyFile: privateKeyFile,
			BackupsDir:     backupsDir,
		},
	})
	if err != nil {
		fmt.Printf("unable to write copier conf. err=%v\n", err)
		return
	}
}

func (p *configurePrompt) runDirectoryConfigureCommand(args []string) {
}

func (p *configurePrompt) runCheckFrequencyConfigureCommand(args []string) {
}

func (p *configurePrompt) runCleanupFrequencyConfigureCommand(args []string) {
}

func (p *configurePrompt) runDateFormatConfigureCommand(args []string) {
}
