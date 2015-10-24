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

	copiers  map[string]*scpRemoteCopier
	cleaners map[string]*cleaner
}

func newEngine(logger *logger) *engine {
	return &engine{
		logger:  logger,
		stopCh:  make(chan struct{}),
		copiers: make(map[string]*scpRemoteCopier),
	}
}

func (e *engine) init() error {
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

	// Close and clean the existing cleaners
	{
		for _, v := range e.cleaners {
			v.Close()
		}
		e.cleaners = make(map[string]*cleaner)
	}

	for _, c := range e.conf.SCPCopiers {
		params := c.Params

		{
			cop := newSCPRemoteCopier(e.logger, c.Name, &params)

			if err := cop.client.connect(); err != nil {
				e.logger.Infof(1, "unable to connect copier %s. err=%v", c, err)
				continue
			}

			e.copiers[c.Name] = cop
		}

		{
			cleaner := newCleaner(e.logger, &params)

			if err := cleaner.client.connect(); err != nil {
				e.logger.Infof(1, "unable to connect cleaner %s. err=%v", c, err)
				continue
			}

			e.cleaners[c.Name] = cleaner
		}
	}

	return nil
}

func fanInWithImmediateStart(in <-chan time.Time) <-chan struct{} {
	forceCh := make(chan struct{}, 1)
	forceCh <- struct{}{}

	ch := make(chan struct{})
	go func() {
		for {
			select {
			case <-in:
				ch <- struct{}{}
			case <-forceCh:
				ch <- struct{}{}
			}
		}
	}()
	return ch
}

func (e *engine) run() {
	if err := e.readConfig(); err != nil {
		e.logger.Errorf(1, "unable to read configuration, bailing. err=%v", err)
		return
	}

	backupTicker := time.NewTicker(e.conf.CheckFrequency.Duration)
	cleanupTicker := time.NewTicker(e.conf.CleanupFrequency.Duration)

	backupCh := fanInWithImmediateStart(backupTicker.C)
	cleanupCh := fanInWithImmediateStart(cleanupTicker.C)

loop:
	for {
		select {
		case <-backupCh:
			e.logger.Infof(1, "starting backup")

			ood, err := e.getOutOfDate()
			if err != nil {
				e.logger.Errorf(1, "unable to get out of date backups. err=%v", err)
				continue loop
			}

			for _, id := range ood {
				e.backupOne(id)
			}

			e.logger.Infof(1, "backup done")

		case <-cleanupCh:
			e.logger.Infof(1, "starting cleanup")

			directories, err := e.getExpirable()
			if err != nil {
				e.logger.Errorf(1, "unable to get expirable backups. err=%v", err)
				continue loop
			}

			for _, id := range directories {
				e.expireOne(id)
			}

			e.logger.Infof(1, "cleanup done")

		case <-e.stopCh:
			break loop
		}
	}
}

func (e *engine) backupOne(id *directory) {
	// First init the copiers because the user might have added copiers to the config file.
	if err := e.init(); err != nil {
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

	name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().UTC().Format(e.conf.DateFormat))

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

func (e *engine) expireOne(id *directory) {
	if err := e.init(); err != nil {
		e.logger.Errorf(1, "unable to init copiers. err=%v", err)
		return
	}

	for _, cleaner := range e.cleaners {
		if err := cleaner.cleanAllExpiredBackups(id, e.conf.DateFormat); err != nil {
			e.logger.Errorf(1, "unable to clean all expired backups. err=%v", err)
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

func (e *engine) getExpirable() (res []*directory, err error) {
	if err := e.readConfig(); err != nil {
		return nil, err
	}

	for _, d := range e.conf.Directories {
		if d.MaxBackups > 0 || d.MaxBackupAge.Duration > 0 {
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
