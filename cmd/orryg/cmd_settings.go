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

func settingsUsageError() error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Usage: orryg settings [options] <subcommand> [arguments]\n\n")
	fmt.Fprintf(&buf, "  Available sub commands\n\n")
	fmt.Fprintf(&buf, "%20s   %s\n", "list", "List all settings")
	fmt.Fprintf(&buf, "%20s   %s\n", "change", "Change a setting")

	return errors.New(buf.String())
}

func settingsCommand(args ...string) error {
	if len(args) < 1 {
		return settingsUsageError()
	}

	remainingArgs := args[1:]

	switch v := strings.ToLower(args[0]); v {
	case "ls", "list":
		return settingsListCommand(remainingArgs...)
	case "change":
		return settingsChangeCommand(remainingArgs...)
	default:
		return fmt.Errorf("unknown settings subcommand '%s'", v)
	}
}

func settingsListCommand(args ...string) error {
	var s orryg.Settings

	err := cl.postAndUnmarshal("/settings/list", nil, &s)
	if err != nil {
		return err
	}

	fmt.Println("\nSettings\n")

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Key", "Value"})

	table.Append([]string{"checkFrequency", s.CheckFrequency.String()})
	table.Append([]string{"dateFormat", s.DateFormat})

	table.Render()
	fmt.Println("")

	return nil
}

func settingsChangeCommand(args ...string) error {
	if len(args) < 2 {
		return errors.New("Usage: orryg settings change <key> <value>")
	}

	key := args[0]
	value := args[1]

	var se orryg.Settings
	switch strings.ToLower(key) {
	case "checkfrequency":
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}

		se.CheckFrequency = d
	case "dateformat":
		se.DateFormat = value
	}

	s, err := cl.post("/settings/change", se)
	if err != nil {
		return err
	}

	fmt.Println(s)

	return nil
}
