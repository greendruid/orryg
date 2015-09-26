package main

import (
	"fmt"
	"strings"
)

func copiersCommand(args ...string) error {
	if len(args) <= 1 {
		return fmt.Errorf("not enough arguments")
	}

	remainingArgs := args[1:]

	switch strings.ToLower(args[0]) {
	case "list":
		return copiersListCommand(remainingArgs...)
	case "add":
		return copiersAddCommand(remainingArgs...)
	case "remove":
		return copiersRemoveCommand(remainingArgs...)
	}
	return nil
}

func copiersListCommand(args ...string) error {
	return nil
}

func copiersAddCommand(args ...string) error {
	return nil
}

func copiersRemoveCommand(args ...string) error {
	return nil
}
