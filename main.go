package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/vrischmann/userdir"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

type service struct {
	logger *logger
	e      *engine
}

func (s *service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	changes <- svc.Status{
		State:   svc.Running,
		Accepts: cmdsAccepted,
	}

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus

			case svc.Stop, svc.Shutdown:
				s.logger.Infof(1, "stopping")

				if err := s.e.stop(); err != nil {
					s.logger.Errorf(1, err.Error())
				}

				break loop

			default:
				s.logger.Errorf(1, "unexpected control request #%d", c)
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}

	return
}

func getLogFile(elog debug.Log) io.Writer {
	dir := filepath.Join(userdir.GetDataHome(), "orryg")

	{
		fi, err := os.Stat(dir)
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(dir, 0700)
		} else if err != nil && !os.IsNotExist(err) {
			elog.Warning(1, fmt.Sprintf("unable to create log directory %s. err=%v", dir, err))
			return ioutil.Discard
		} else {
			if !fi.IsDir() {
				elog.Warning(1, fmt.Sprintf("unable to create log directory %s because it's already a file", dir))
				return ioutil.Discard
			}
		}
	}

	file := filepath.Join(dir, "main.log")
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		elog.Warning(1, fmt.Sprintf("unable to create log file %s. err=%v", file, err))
		return ioutil.Discard
	}

	return f
}

func runService(name string, isDebug bool) {
	var elog debug.Log
	var err error
	{
		if isDebug {
			elog = debug.New(name)
		} else {
			elog, err = eventlog.Open(name)
			if err != nil {
				return
			}
		}
		defer elog.Close()
	}

	logger := &logger{
		elog:   elog,
		stdLog: log.New(getLogFile(elog), "orryg: ", log.LstdFlags),
	}

	{
		_, err = os.Stat(configPath)
		if err != nil && os.IsNotExist(err) {
			logger.Errorf(1, "configuration file at %s does not exist, please create it", configPath)
			return
		} else if err != nil {
			logger.Errorf(1, "unable to read configuration file at %s. err=%v", configPath, err)
			return
		}
	}

	var run func(string, svc.Handler) error
	{
		run = svc.Run
		if isDebug {
			run = debug.Run
		}
	}

	logger.Infof(1, "starting %s service", name)

	var sv *service
	{
		var e *engine
		{
			logger.Infof(1, "starting engine")

			var err error
			e, err = newEngine(logger)
			if err != nil {
				logger.Errorf(1, err.Error())
				return
			}
			go e.run()

			logger.Infof(1, "engine started")
		}

		sv = &service{
			e:      e,
			logger: logger,
		}
	}

	err = run(name, sv)
	if err != nil {
		logger.Errorf(1, "%s service failed: %v", name, err)
		return
	}

	logger.Infof(1, "%s service stopped", name)
}

func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}

	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}

	timeout := time.Now().Add(10 * time.Second)

	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}

		time.Sleep(300 * time.Millisecond)

		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}

	return nil
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}

	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}

	return "", err
}

func installService(name, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: desc}, "is", "auto-started")
	if err != nil {
		return err
	}
	defer s.Close()

	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}

	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()

	err = s.Delete()
	if err != nil {
		return err
	}

	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}

	return nil
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop.\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	go http.ListenAndServe(":6060", nil)

	const svcName = "Orryg"

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}

	if !isIntSess {
		runService(svcName, false)
		return
	}

	if len(os.Args) < 2 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "debug":
		configPath = filepath.Join(userdir.GetConfigHome(), "orryg", "config.json")
		runService(svcName, true)
		return
	case "install":
		err = installService(svcName, "Backup service")
	case "remove":
		err = removeService(svcName)
	case "start":
		err = startService(svcName)
	case "stop":
		err = controlService(svcName, svc.Stop, svc.Stopped)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}

	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, svcName, err)
	}
}
