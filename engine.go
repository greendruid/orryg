package main

import (
	"fmt"
	"time"
)

type engine struct {
	conf configuration

	stopCh chan struct{}
	doneCh chan struct{}

	copiers  map[string]*scpRemoteCopier
	cleaners map[string]*cleaner
}

func newEngine(conf configuration) *engine {
	return &engine{
		conf:    conf,
		stopCh:  make(chan struct{}),
		doneCh:  make(chan struct{}),
		copiers: make(map[string]*scpRemoteCopier),
	}
}

func (e *engine) init() error {
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

	copiers, err := e.conf.ReadSCPCopiers()
	if err != nil {
		return fmt.Errorf("unable to read copiers configuration. err=%v", err)
	}
	for _, c := range copiers {
		params := c.Params

		{
			cop := newSCPRemoteCopier(c.Name, &params)

			if err := cop.client.connect(); err != nil {
				logger.Printf("unable to connect copier %s. err=%v", c, err)
				continue
			}

			e.copiers[c.Name] = cop
		}

		{
			cleaner := newCleaner(&params)

			if err := cleaner.client.connect(); err != nil {
				logger.Printf("unable to connect cleaner %s. err=%v", c, err)
				continue
			}

			e.cleaners[c.Name] = cleaner
		}
	}

	return nil
}

func (e *engine) stop() error {
	for _, c := range e.copiers {
		if err := c.Close(); err != nil {
			return err
		}
	}

	e.stopCh <- struct{}{}

	<-e.doneCh

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
	checkFrequency, err := e.conf.ReadCheckFrequency()
	if err != nil {
		logger.Printf("unable to read check frequency, using default value of 1 minute. err=%v", err)
		checkFrequency = time.Minute
	}

	cleanupFrequency, err := e.conf.ReadCleanupFrequency()
	if err != nil {
		logger.Printf("unable to read cleanup frequency, using default value of 1 minute. err=%v", err)
		cleanupFrequency = time.Minute
	}

	backupTicker := time.NewTicker(checkFrequency)
	cleanupTicker := time.NewTicker(cleanupFrequency)

	backupCh := fanInWithImmediateStart(backupTicker.C)
	cleanupCh := fanInWithImmediateStart(cleanupTicker.C)

	// keep a big margin just to be safe
	backupOneCh := make(chan directory, 1024)
	cleanupOneCh := make(chan directory, 1024)

loop:
	for {
		select {
		case <-backupCh:
			ood, err := e.getOutOfDate()
			if err != nil {
				logger.Printf("unable to get out of date backups. err=%v", err)
				continue loop
			}

			for _, id := range ood {
				backupOneCh <- id
			}

		case id := <-backupOneCh:
			e.backupOne(id)

		case <-cleanupCh:
			directories, err := e.getExpirable()
			if err != nil {
				logger.Printf("unable to get expirable backups. err=%v", err)
				continue loop
			}

			for _, id := range directories {
				cleanupOneCh <- id
			}

		case id := <-cleanupOneCh:
			e.expireOne(id)

		case <-e.stopCh:
			break loop
		}
	}

	e.doneCh <- struct{}{}
}

func (e *engine) backupOne(id directory) {
	// First init the copiers because the user might have added copiers to the config file.
	if err := e.init(); err != nil {
		logger.Printf("unable to init copiers. err=%v", err)
		return
	}

	start := time.Now()

	logger.Printf("backing up %s", id.OriginalPath)

	logger.Printf("making tarball of %s", id.OriginalPath)
	tb := newTarball(id)
	defer func() {
		// Cleanup the tarball
		if err := tb.Close(); err != nil {
			logger.Printf("unable to close tarball. err=%v", err)
		}
	}()

	if err := tb.process(); err != nil {
		logger.Printf("unable to make tarball. err=%v", err)
		return
	}
	logger.Printf("done making tarball of %s", id.OriginalPath)

	dateFormat, err := e.conf.ReadDateFormat()
	if err != nil {
		logger.Printf("unable to read date format. err=%v", err)
		return
	}

	name := fmt.Sprintf("%s_%s.tar.gz", id.ArchiveName, time.Now().UTC().Format(dateFormat))

	for _, copier := range e.copiers {
		tb.Reset()

		logger.Printf("start copying %s with copier %s", name, copier)

		err := copier.CopyFromReader(tb, tb.fi.Size(), name)
		if err != nil {
			logger.Printf("unable to copy the tarball to the remote host. err=%v", err)
			return // don't continue if even one copier is not working.
		}

		logger.Printf("done copying %s with copier %s", name, copier)
	}

	logger.Printf("backed up %s in %s", id.OriginalPath, time.Now().Sub(start))

	// Persist the new config
	{
		if err := e.conf.UpdateLastUpdated(id); err != nil {
			logger.Printf("unable to update the last updated field of %v. err=%v", id, err)
		}
	}
}

func (e *engine) expireOne(id directory) {
	if err := e.init(); err != nil {
		logger.Printf("unable to init copiers. err=%v", err)
		return
	}

	dateFormat, err := e.conf.ReadDateFormat()
	if err != nil {
		logger.Printf("unable to read date format. err=%v", err)
		return
	}

	for _, cleaner := range e.cleaners {
		if err := cleaner.cleanAllExpiredBackups(id, dateFormat); err != nil {
			logger.Printf("unable to clean all expired backups. err=%v", err)
		}
	}
}

func (e *engine) getOutOfDate() (res []directory, err error) {
	directories, err := e.conf.ReadDirectories()
	if err != nil {
		return nil, fmt.Errorf("unable to read directories configuration. err=%v", err)
	}

	for _, d := range directories {
		elapsed := time.Now().Sub(d.LastUpdated)
		if d.LastUpdated.IsZero() || elapsed >= d.Frequency {
			res = append(res, d)
		}
	}

	return
}

func (e *engine) getExpirable() (res []directory, err error) {
	directories, err := e.conf.ReadDirectories()
	if err != nil {
		return nil, fmt.Errorf("unable to read directories configuration. err=%v", err)
	}

	for _, d := range directories {
		if d.MaxBackups > 0 || d.MaxBackupAge > 0 {
			res = append(res, d)
		}
	}

	return
}
