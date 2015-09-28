package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/vrischmann/orryg"
)

func directoriesUsageError() error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Usage: orryg directories [options] <subcommand> [arguments]\n\n")
	fmt.Fprintf(&buf, "  Available sub commands\n\n")
	fmt.Fprintf(&buf, "%20s   %s\n", "list", "List all directories")
	fmt.Fprintf(&buf, "%20s   %s\n", "add", "Add a directory")
	fmt.Fprintf(&buf, "%20s   %s\n", "remove", "Remove a directory")

	return errors.New(buf.String())
}

func directoriesCommand(args ...string) error {
	if len(args) < 1 {
		return directoriesUsageError()
	}

	remainingArgs := args[1:]

	switch v := strings.ToLower(args[0]); v {
	case "ls", "list":
		return directoriesListCommand(remainingArgs...)
	case "add":
		return directoriesAddCommand(remainingArgs...)
	case "rm", "remove":
		return directoriesRemoveCommand(remainingArgs...)
	default:
		return fmt.Errorf("unknown directories subcommand '%s'", v)
	}
}

func directoriesListCommand(args ...string) error {
	var directories []orryg.Directory

	err := cl.postAndUnmarshal("/directories/list", nil, &directories)
	if err != nil {
		return err
	}

	if len(directories) > 0 {
		fmt.Println("\nDirectories\n")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Original path", "Archive name", "Frequency", "Last backup", "Next backup"})

		for _, d := range directories {
			lastBackup := "Never backed up"
			nextBackup := ""
			if !d.LastUpdated.IsZero() {
				lastBackup = d.LastUpdated.Format("2006-01-02 15:04")
				nextBackup = d.LastUpdated.Add(d.Frequency).Format("2006-01-02 15:04")
			}

			table.Append([]string{
				d.OriginalPath, d.ArchiveName, d.Frequency.String(),
				lastBackup, nextBackup,
			})
		}

		table.Render()
		fmt.Println("")
	}

	return nil
}

func directoriesAddCommand(args ...string) error {
	ip := newInput()

	originalPath := ip.read("Path? ")[0]
	archiveName := ip.read("Archive name? ")[0]
	frequency := ip.read("Frequency of backup? (10m, 6h, etc) ")[0]

	if err := ip.Close(); err != nil {
		return err
	}

	freq, err := time.ParseDuration(frequency)
	if err != nil {
		return err
	}

	if archiveName == "" {
		return errors.New("please provide an archive name")
	}

	_, err = os.Stat(originalPath)
	if err != nil {
		return fmt.Errorf("%s is not valid. %v", originalPath, err)
	}

	req := orryg.Directory{
		Frequency:    freq,
		OriginalPath: originalPath,
		ArchiveName:  archiveName,
	}

	s, err := cl.post("/directories/add", req)
	if err != nil {
		return err
	}

	fmt.Println(s)

	return nil
}

func directoriesRemoveCommand(args ...string) error {
	return nil
}
