package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/vrischmann/jsonutil"
	"github.com/vrischmann/userdir"
)

type copierType uint

const (
	unknownCopierType copierType = iota
	scpCopierType
)

func (t copierType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *copierType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	val, err := newCopierTypeFromString(s)
	if err != nil {
		return err
	}

	*t = val

	return nil
}

func (t copierType) String() string {
	switch t {
	case scpCopierType:
		return "scp"
	default:
		return "unknown"
	}
}

func newCopierTypeFromString(s string) (copierType, error) {
	switch strings.ToLower(s) {
	case "scp":
		return scpCopierType, nil
	default:
		return unknownCopierType, fmt.Errorf("unknown copier type %s", s)
	}
}

type copierConf struct {
	Type copierType      `json:"type"`
	Conf json.RawMessage `json:"conf"`
}

type directoryConf struct {
	Frequency   jsonutil.Duration `json:"frequency"`
	OrigPath    string            `json:"origPath"`
	ArchiveName string            `json:"archiveName"`
}

type internalDirectory struct {
	Frequency   time.Duration
	OrigPath    string
	ArchiveName string
	LastUpdated time.Time
}

func (d *internalDirectory) merge(id internalDirectory) {
	d.Frequency = id.Frequency
	d.OrigPath = id.OrigPath
	d.ArchiveName = id.ArchiveName
	d.LastUpdated = id.LastUpdated
}

var conf struct {
	CheckFrequency jsonutil.Duration `json:"checkFrequency"`
	DateFormat     string            `json:"dateFormat"`
	Copiers        []copierConf      `json:"copiers"`
	Directories    []directoryConf   `json:"directories"`
}

var (
	copiers []remoteCopier
	sched   *scheduler
)

func readConfig() error {
	path := filepath.Join(userdir.GetConfigHome(), "orryg", "config.json")

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)
	err = dec.Decode(&conf)
	if err != nil {
		return err
	}

	if conf.CheckFrequency.Duration == 0 {
		conf.CheckFrequency.Duration = time.Minute * 1
	}

	return nil
}

func main() {
	{
		err := readConfig()
		if err != nil {
			log.Fatalln(err)
		}

		sched, err = newScheduler()
		if err != nil {
			log.Fatalln(err)
		}

		err = sched.init()
		if err != nil {
			log.Fatalln(err)
		}
	}

	go sched.run()

	for _, c := range conf.Copiers {
		switch c.Type {
		case scpCopierType:
			var params sshParameters

			if err := json.Unmarshal(c.Conf, &params); err != nil {
				log.Fatalln(err)
			}

			cop := newSCPRemoteCopier(&params)
			defer cop.Close()

			if err := cop.Connect(); err != nil {
				log.Fatalln(err)
			}

			copiers = append(copiers, cop)
		}
	}

	for _, c := range conf.Directories {
		if err := sched.mergeDirectory(c); err != nil {
			log.Fatalf("unable to merge directory %+v. err=%v", c, err)
		}
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Kill, os.Interrupt)

loop:
	for {
		select {
		case <-signalCh:
			log.Printf("stopping")
			break loop
		case id := <-sched.ch:
			start := time.Now()

			log.Printf("backing up %s", id.OrigPath)

			tb := newTarball(id)
			if err := tb.process(); err != nil {
				log.Fatalf("unable to make tarball. err=%v", err)
			}

			name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().Format(conf.DateFormat))

			for _, copier := range copiers {
				err := copier.CopyFromReader(tb, tb.fi.Size(), name)
				if err != nil {
					log.Fatalf("unable to copy the tarball to the remote host. err=%v", err)
				}
			}

			if err := tb.Close(); err != nil {
				log.Fatalf("unable to close tarball. err=%v", err)
			}

			elapsed := time.Now().Sub(start)

			log.Printf("backed up %s in %s", id.OrigPath, elapsed)

			id.LastUpdated = time.Now()
			if err := sched.updateDirectory(id); err != nil {
				log.Fatalf("unable to merge directory %+v. err=%v", id, err)
			}
		}
	}
}
