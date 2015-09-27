package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/peterh/liner"
	"github.com/vrischmann/orryg"
	"github.com/vrischmann/shlex"
)

func copiersCommand(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments")
	}

	remainingArgs := args[1:]

	switch strings.ToLower(args[0]) {
	case "list":
		return copiersListCommand(remainingArgs...)
	case "add-scp":
		return copiersAddSCPCommand(remainingArgs...)
	case "remove":
		return copiersRemoveCommand(remainingArgs...)
	}
	return nil
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
		fmt.Println()
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
	return nil
}

type input struct {
	line *liner.State
	err  error
}

func newInput() *input {
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)

	return &input{line: line}
}

func (i *input) read(prompt string) (res []string) {
	if i.err != nil {
		return []string{""}
	}

	var cmd string
	cmd, i.err = i.line.Prompt(prompt)
	if i.err != nil {
		return []string{""}
	}

	return shlex.Parse(cmd)
}

func (i *input) Close() error {
	i.line.Close()
	return i.err
}
