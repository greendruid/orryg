package main

import (
	"github.com/peterh/liner"
	"github.com/vrischmann/shlex"
)

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
