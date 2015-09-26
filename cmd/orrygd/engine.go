package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vrischmann/orryg"
)

type engine struct {
	st     *dataStore
	oodCh  chan orryg.Directory
	stopCh chan struct{}

	copiers []remoteCopier
}

func newEngine(st *dataStore) (*engine, error) {
	e := &engine{
		st:     st,
		oodCh:  make(chan orryg.Directory),
		stopCh: make(chan struct{}),
	}

	if err := e.initCopiers(); err != nil {
		return nil, err
	}

	go e.scheduleOOD()

	return e, nil
}

func (e *engine) initCopiers() error {
	confs, err := e.st.getAllSCPCopierConfs()
	if err != nil {
		return err
	}

	for _, c := range confs {
		cop := newSCPRemoteCopier(&c.Params)

		if err := cop.Connect(); err != nil {
			return err
		}

		e.copiers = append(e.copiers, cop)
	}

	return nil
}

func (e *engine) run() {
loop:
	for {
		select {
		case <-e.stopCh:
			break loop
		case id := <-e.oodCh:
			start := time.Now()

			log.Printf("backing up %s", id.OriginalPath)

			tb := newTarball(id)
			if err := tb.process(); err != nil {
				log.Printf("unable to make tarball. err=%v", err)
				continue
			}

			settings, err := e.st.getSettings()
			if err != nil {
				log.Printf("unable to get settings. err=%v", err)
				continue
			}

			name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().Format(settings.DateFormat))

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

			id.LastUpdated = time.Now()
			if err := e.st.mergeDirectory(id); err != nil {
				log.Fatalf("unable to merge directory %+v. err=%v", id, err)
			}
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
	settings, err := e.st.getSettings()
	if err != nil {
		log.Printf("unable to get settings. err=%v", err)
		return
	}

	ticker := time.NewTicker(settings.CheckFrequency)

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

func (e *engine) getOutOfDate() (res []orryg.Directory, err error) {
	err = e.st.forEeachDirectory(func(d orryg.Directory) error {
		elapsed := time.Now().Sub(d.LastUpdated)
		if d.LastUpdated.IsZero() || elapsed >= d.Frequency {
			res = append(res, d)
		}

		return nil
	})
	return
}
