package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		default:
			fmt.Printf("invalid command '%s'\n", cmd)
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
	if len(args) == 0 {
		directories, err := p.conf.ReadDirectories()
		if err != nil {
			fmt.Printf("unable to read directories. err=%v\n", err)
			return
		}

		var buf bytes.Buffer
		fmt.Fprintf(&buf, "directories:\n")
		for _, directory := range directories {
			fmt.Fprintf(&buf, "%s\n", directory)
		}

		fmt.Println(buf.String())

		return
	}

	if len(args) != 5 {
		fmt.Println("Usage: directory <original path> <archive name> <frequency> <max backups> <max backup age>")
		return
	}

	originalPath := args[0]
	archiveName := args[1]
	sFrequency := args[2]
	sMaxBackups := args[3]
	sMaxBackupAge := args[4]

	frequency, err := time.ParseDuration(sFrequency)
	if err != nil {
		fmt.Printf("frequency '%s' is not valid\n", sFrequency)
		return
	}

	maxBackups, err := strconv.Atoi(sMaxBackups)
	if err != nil {
		fmt.Printf("max backups number '%s' is not valid\n", sMaxBackupAge)
		return
	}

	maxBackupAge, err := time.ParseDuration(sMaxBackupAge)
	if err != nil {
		fmt.Printf("max backup age '%s' is not valid\n", sMaxBackupAge)
		return
	}

	err = p.conf.WriteDirectory(directory{
		Frequency:    frequency,
		OriginalPath: originalPath,
		ArchiveName:  archiveName,
		MaxBackups:   maxBackups,
		MaxBackupAge: maxBackupAge,
	})
	if err != nil {
		fmt.Printf("unable to write directory conf. err=%v\n", err)
		return
	}
}

func (p *configurePrompt) runCheckFrequencyConfigureCommand(args []string) {
	if len(args) == 0 {
		freq, err := p.conf.ReadCheckFrequency()
		if err != nil {
			fmt.Printf("unable to read check frequency. err=%v\n", err)
			return
		}

		fmt.Println(freq)

		return
	}

	if len(args) != 1 {
		fmt.Println("Usage: check-frequency <frequency>")
		return
	}

	frequency, err := time.ParseDuration(args[0])
	if err != nil {
		fmt.Printf("frequency '%s' is not valid\n", args[0])
		return
	}

	err = p.conf.WriteCheckFrequency(frequency)
	if err != nil {
		fmt.Printf("unable to write check frequency conf. err=%v\n", err)
		return
	}
}

func (p *configurePrompt) runCleanupFrequencyConfigureCommand(args []string) {
	if len(args) == 0 {
		freq, err := p.conf.ReadCleanupFrequency()
		if err != nil {
			fmt.Printf("unable to read cleanup frequency. err=%v\n", err)
			return
		}

		fmt.Println(freq)

		return
	}

	if len(args) != 1 {
		fmt.Println("Usage: cleanup-frequency <frequency>")
		return
	}

	frequency, err := time.ParseDuration(args[0])
	if err != nil {
		fmt.Printf("frequency '%s' is not valid\n", args[0])
		return
	}

	err = p.conf.WriteCleanupFrequency(frequency)
	if err != nil {
		fmt.Printf("unable to write cleanup frequency conf. err=%v\n", err)
		return
	}
}

func (p *configurePrompt) runDateFormatConfigureCommand(args []string) {
	if len(args) == 0 {
		format, err := p.conf.ReadDateFormat()
		if err != nil {
			fmt.Printf("unable to read date format. err=%v\n", err)
			return
		}

		fmt.Println(format)

		return
	}

	if len(args) != 1 {
		fmt.Println("Usage: date-format <format>")
		return
	}

	format := args[0]

	// Verify the validity of the format
	t1 := time.Now().Format(format)
	if t, err := time.Parse(format, t1); err != nil || t.IsZero() {
		fmt.Printf("invalid date format '%s'", format)
		return
	}

	err := p.conf.WriteDateFormat(format)
	if err != nil {
		fmt.Printf("unable to write date format conf. err=%v\n", err)
		return
	}
}
