package main

import (
	"fmt"
	"log"
	"time"
)

type engine struct {
	oodCh  chan *directory
	stopCh chan struct{}

	copiers map[string]remoteCopier
}

func newEngine() (*engine, error) {
	e := &engine{
		oodCh:   make(chan *directory),
		stopCh:  make(chan struct{}),
		copiers: make(map[string]remoteCopier),
	}

	if err := e.initCopiers(); err != nil {
		return nil, err
	}

	go e.scheduleOOD()

	return e, nil
}

func (e *engine) initCopiers() error {
	if err := readConfig(); err != nil {
		return err
	}

	for _, c := range conf.SCPCopiers {
		cop := newSCPRemoteCopier(&c.Params)

		if err := cop.Connect(); err != nil {
			log.Printf("unable to connect copier %s. err=%v", c.Name, err)
			continue
		}

		e.copiers[c.Name] = cop
	}

	return nil
}

func (e *engine) run() {
loop:
	for {
		select {
		case id := <-e.oodCh:
			start := time.Now()

			log.Printf("backing up %s", id.OriginalPath)

			tb := newTarball(id)
			if err := tb.process(); err != nil {
				log.Printf("unable to make tarball. err=%v", err)
				continue
			}

			name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().Format(conf.DateFormat))

			for _, copier := range e.copiers {
				err := copier.CopyFromReader(tb, tb.fi.Size(), name)
				if err != nil {
					log.Fatalf("unable to copy the tarball to the remote host. err=%v", err)
				}
			}

			if err := tb.Close(); err != nil {
				log.Fatalf("unable to close tarball. err=%v", err)
			}

			elapsed := time.Now().Sub(start)

			log.Printf("backed up %s in %s", id.OriginalPath, elapsed)

			// conf.setLastUpdated(id, time.Now())
			id.LastUpdated = time.Now()

			if err := writeConfig(); err != nil {
				log.Fatalf("unable to write config. err=%v", err)
			}

			log.Printf("id: %v", id)

		case <-e.stopCh:
			break loop
		}
	}
}

func (e *engine) stop() error {
	for _, c := range e.copiers {
		if err := c.Close(); err != nil {
			return err
		}
	}

	e.stopCh <- struct{}{}
	e.stopCh <- struct{}{}

	return nil
}

func (e *engine) scheduleOOD() {
	ticker := time.NewTicker(conf.CheckFrequency.Duration)

loop:
	for {
		select {
		case <-ticker.C:
			ood, err := e.getOutOfDate()
			if err != nil {
				log.Printf("unable to get out of date backups. err=%v", err)
				continue
			}

			for _, d := range ood {
				e.oodCh <- d
			}
		case <-e.stopCh:
			break loop
		}
	}
}

func (e *engine) getOutOfDate() (res []*directory, err error) {
	if err := readConfig(); err != nil {
		return nil, err
	}

	for _, d := range conf.Directories {
		elapsed := time.Now().Sub(d.LastUpdated)
		if d.LastUpdated.IsZero() || elapsed >= d.Frequency.Duration {
			res = append(res, d)
		}
	}

	return
}
