package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/vrischmann/orryg"
)

func copiersUsageError() error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Usage: orryg copiers [options] <subcommand> [arguments]\n\n")
	fmt.Fprintf(&buf, "  Available sub commands\n\n")
	fmt.Fprintf(&buf, "%20s   %s\n", "list", "List all copiers")
	fmt.Fprintf(&buf, "%20s   %s\n", "add", "Add a copier")
	fmt.Fprintf(&buf, "%20s   %s\n", "remove", "Remove a copier")

	return errors.New(buf.String())
}

func copiersCommand(args ...string) error {
	if len(args) < 1 {
		return copiersUsageError()
	}

	remainingArgs := args[1:]

	switch v := strings.ToLower(args[0]); v {
	case "ls", "list":
		return copiersListCommand(remainingArgs...)
	case "add-scp":
		return copiersAddSCPCommand(remainingArgs...)
	case "rm", "remove":
		return copiersRemoveCommand(remainingArgs...)
	default:
		return fmt.Errorf("unknown copiers subcommand '%s'", v)
	}
}

func copiersListCommand(args ...string) error {
	var confs []orryg.UCopierConf

	err := cl.postAndUnmarshal("/copiers/list", nil, &confs)
	if err != nil {
		return err
	}

	var scpConfs []orryg.SCPCopierConf

	for _, conf := range confs {
		switch conf.Type {
		case orryg.SCPCopierType:
			var scpConf orryg.SCPCopierConf
			err := json.Unmarshal(conf.Conf, &scpConf)
			if err != nil {
				return err
			}
			scpConfs = append(scpConfs, scpConf)
		}
	}

	if len(scpConfs) > 0 {
		fmt.Println("\nSCP Copiers\n")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "User", "Host", "Port", "Private Key File", "Backups Directory"})

		for _, conf := range scpConfs {
			table.Append([]string{
				conf.Name, conf.Params.User, conf.Params.Host, fmt.Sprintf("%d", conf.Params.Port),
				conf.Params.PrivateKeyFile, conf.Params.BackupsDir,
			})
		}

		table.Render()
		fmt.Println("")
	}

	return nil
}

func copiersAddSCPCommand(args ...string) error {
	ip := newInput()

	name := ip.read("Name? ")[0]
	user := ip.read("SSH user? ")[0]
	host := ip.read("SSH host? ")[0]
	port := ip.read("SSH port? ")[0]
	privateKeyFile := ip.read("SSH private key file? ")[0]
	backupsDir := ip.read("SSH backups directory? ")[0]

	if err := ip.Close(); err != nil {
		return err
	}

	tmp, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	req := orryg.CopierConf{
		Type: orryg.SCPCopierType,
		Conf: orryg.SCPCopierConf{
			Name: name,
			Params: orryg.SSHParameters{
				User:           user,
				Host:           host,
				Port:           tmp,
				PrivateKeyFile: privateKeyFile,
				BackupsDir:     backupsDir,
			},
		},
	}

	s, err := cl.post("/copiers/add", req)
	if err != nil {
		return err
	}

	fmt.Println(s)

	return nil
}

func copiersRemoveCommand(args ...string) error {
	if len(args) < 1 {
		return errors.New("Usage: orryg directories remove <name>")
	}

	name := args[0]

	s, err := cl.post("/copiers/remove/"+name, nil)
	if err != nil {
		return err
	}

	fmt.Println(s)

	return nil
}
