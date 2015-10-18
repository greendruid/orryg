package main

import (
	"fmt"

	"golang.org/x/sys/windows/svc/debug"
)

type logger struct {
	elog debug.Log
}

func (l *logger) Infof(eid uint32, format string, args ...interface{}) error {
	return l.elog.Info(eid, fmt.Sprintf(format, args...))
}

func (l *logger) Warnf(eid uint32, format string, args ...interface{}) error {
	return l.elog.Warning(eid, fmt.Sprintf(format, args...))
}

func (l *logger) Errorf(eid uint32, format string, args ...interface{}) error {
	return l.elog.Error(eid, fmt.Sprintf(format, args...))
}
