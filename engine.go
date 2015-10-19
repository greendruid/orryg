package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type engine struct {
	logger *logger
	conf   config

	stopCh chan struct{}

	copiers map[string]*scpRemoteCopier
}

func newEngine(logger *logger) *engine {
	return &engine{
		logger:  logger,
		stopCh:  make(chan struct{}),
		copiers: make(map[string]*scpRemoteCopier),
	}
}

func (e *engine) initCopiers() error {
	e.logger.Infof(1, "initializing copiers")

	if err := e.readConfig(); err != nil {
		return err
	}

	// Close and clean the existing copiers
	{
		for _, v := range e.copiers {
			v.Close()
		}
		e.copiers = make(map[string]*scpRemoteCopier)
	}

	for _, c := range e.conf.SCPCopiers {
		params := c.Params
		cop := newSCPRemoteCopier(e.logger, c.Name, &params)

		if err := cop.Connect(); err != nil {
			e.logger.Infof(1, "unable to connect copier %s. err=%v", c, err)
			continue
		}

		e.copiers[c.Name] = cop
	}

	return nil
}

func (e *engine) run() {
	if err := e.readConfig(); err != nil {
		e.logger.Errorf(1, "unable to read configuration, bailing. err=%v", err)
		return
	}

	ticker := time.NewTicker(e.conf.CheckFrequency.Duration)

	forceCh := make(chan struct{}, 1)
	forceCh <- struct{}{}

	ch := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				ch <- struct{}{}
			case <-forceCh:
				ch <- struct{}{}
			}
		}
	}()

loop:
	for {
		select {
		case <-ch:
			ood, err := e.getOutOfDate()
			if err != nil {
				e.logger.Infof(1, "unable to get out of date backups. err=%v", err)
				continue loop
			}

			for _, id := range ood {
				e.backupOne(id)
			}

		case <-e.stopCh:
			break loop
		}
	}
}

func (e *engine) backupOne(id *directory) {
	// First init the copiers because the user might have added copiers to the config file.
	if err := e.initCopiers(); err != nil {
		e.logger.Errorf(1, "unable to init copiers. err=%v", err)
		return
	}

	start := time.Now()

	e.logger.Infof(1, "backing up %s", id.OriginalPath)

	e.logger.Infof(1, "making tarball of %s", id.OriginalPath)
	tb := newTarball(id)
	defer func() {
		// Cleanup the tarball
		if err := tb.Close(); err != nil {
			e.logger.Errorf(1, "unable to close tarball. err=%v", err)
		}
	}()

	if err := tb.process(); err != nil {
		e.logger.Infof(1, "unable to make tarball. err=%v", err)
		return
	}
	e.logger.Infof(1, "done making tarball of %s", id.OriginalPath)

	name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().Format(e.conf.DateFormat))

	for _, copier := range e.copiers {
		tb.Reset()

		e.logger.Infof(1, "start copying %s with copier %s", name, copier)

		err := copier.CopyFromReader(tb, tb.fi.Size(), name)
		if err != nil {
			e.logger.Errorf(1, "unable to copy the tarball to the remote host. err=%v", err)
			return // don't continue if even one copier is not working.
		}

		e.logger.Infof(1, "done copying %s with copier %s", name, copier)
	}

	e.logger.Infof(1, "backed up %s in %s", id.OriginalPath, time.Now().Sub(start))

	// Persist the new config
	{
		e.conf.update(id)

		if err := e.writeConfig(); err != nil {
			e.logger.Errorf(1, "unable to write config. err=%v", err)
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

	return nil
}

func (e *engine) getOutOfDate() (res []*directory, err error) {
	if err := e.readConfig(); err != nil {
		return nil, err
	}

	for _, d := range e.conf.Directories {
		elapsed := time.Now().Sub(d.LastUpdated)
		if d.LastUpdated.IsZero() || elapsed >= d.Frequency.Duration {
			res = append(res, d)
		}
	}

	return
}

var configPath = "C:/Windows/System32/config/systemprofile/AppData/Roaming/orryg/config.json"

func (e *engine) readConfig() error {
	e.conf = config{}

	f, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	return dec.Decode(&e.conf)
}

func (e *engine) writeConfig() error {
	f, err := os.OpenFile(configPath, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.MarshalIndent(&e.conf, "", "    ")
	if err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
}
