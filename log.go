package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/svc/debug"
)

type logger struct {
	elog   debug.Log
	stdLog *log.Logger
}

func (l *logger) Infof(eid uint32, format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)

	l.stdLog.Printf(str)
	return l.elog.Info(eid, str)
}

func (l *logger) Warnf(eid uint32, format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)

	l.stdLog.Printf(str)
	return l.elog.Warning(eid, str)
}

func (l *logger) Errorf(eid uint32, format string, args ...interface{}) error {
	str := fmt.Sprintf(format, args...)

	l.stdLog.Printf(str)
	return l.elog.Error(eid, str)
}
