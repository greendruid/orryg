package main

import (
	"encoding/json"
	"log"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/vrischmann/userdir"
)

type scheduler struct {
	db *bolt.DB
	ch chan internalDirectory
}

func newScheduler() (*scheduler, error) {
	path := filepath.Join(userdir.GetDataHome(), "orryg", "orryg.db")

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &scheduler{
		db: db,
		ch: make(chan internalDirectory),
	}, nil
}

var (
	directoriesBucket = []byte("directories")
)

func (s *scheduler) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(directoriesBucket)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *scheduler) run() {
	for range time.Tick(conf.CheckFrequency.Duration) {
		ood, err := s.getOutOfDate()
		if err != nil {
			log.Printf("unable to get out of date backups. err=%v", err)
			continue
		}

		for _, id := range ood {
			s.ch <- id
		}
	}
}

func (s *scheduler) stop() error {
	return s.db.Close()
}

func (s *scheduler) mergeDirectory(conf directoryConf) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		key := []byte(conf.OrigPath)

		var id internalDirectory
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &id); err != nil {
				return err
			}

			id.Frequency = conf.Frequency.Duration
			id.OrigPath = conf.OrigPath
			id.ArchiveName = conf.ArchiveName
		} else {
			id = internalDirectory{
				Frequency:   conf.Frequency.Duration,
				OrigPath:    conf.OrigPath,
				ArchiveName: conf.ArchiveName,
			}
		}

		data, err := json.Marshal(&id)
		if err != nil {
			return err
		}

		return bucket.Put(key, data)
	})
}

func (s *scheduler) updateDirectory(dir internalDirectory) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		key := []byte(dir.OrigPath)

		var id internalDirectory
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &id); err != nil {
				return err
			}

			id.merge(dir)
		} else {
			id = dir
		}

		data, err := json.Marshal(&id)
		if err != nil {
			return err
		}

		return bucket.Put(key, data)
	})
}

func (s *scheduler) getOutOfDate() (res []internalDirectory, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var id internalDirectory

			if err := json.Unmarshal(v, &id); err != nil {
				return err
			}

			elapsed := time.Now().Sub(id.LastUpdated)
			if id.LastUpdated.IsZero() || elapsed >= id.Frequency {
				res = append(res, id)
			}
		}

		return nil
	})
	return
}
