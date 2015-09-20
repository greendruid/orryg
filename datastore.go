package main

import (
	"encoding/json"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/vrischmann/userdir"
)

var (
	copiersBucket     = []byte("copiers")
	directoriesBucket = []byte("directories")
	settingsBucket    = []byte("settings")
)

type dataStore struct {
	db *bolt.DB
}

func newDataStore() (*dataStore, error) {
	path := filepath.Join(userdir.GetDataHome(), "orryg", "orryg.db")

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &dataStore{
		db: db,
	}, nil
}

func (s *dataStore) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		for _, bucket := range [][]byte{copiersBucket, directoriesBucket, settingsBucket} {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *dataStore) Close() error {
	return s.db.Close()
}

func (s *dataStore) getSettings() (se settings, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(settingsBucket)

		data := bucket.Get(settingsBucket)
		if data == nil {
			se = defaultSettings()

			data, err := json.Marshal(se)
			if err != nil {
				return err
			}

			return bucket.Put(settingsBucket, data)
		}

		return json.Unmarshal(data, &se)
	})
	return
}

func (s *dataStore) getAllSCPCopierConfs() (res []scpCopierConf, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(copiersBucket)

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var c scpCopierConf

			if err := json.Unmarshal(v, &c); err != nil {
				return err
			}

			res = append(res, c)
		}

		return nil
	})

	return
}

func (s *dataStore) mergeSCPCopierConf(c scpCopierConf) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(copiersBucket)

		key := append([]byte{byte(scpCopierType)}, []byte(c.Name)...)

		var params sshParameters
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &params); err != nil {
				return err
			}

			params.merge(c.Params)
		} else {
			params = c.Params
		}

		data, err := json.Marshal(&params)
		if err != nil {
			return err
		}

		return bucket.Put(key, data)
	})
}

func (s *dataStore) mergeDirectory(dir directory) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		key := []byte(dir.OrigPath)

		var d directory
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &d); err != nil {
				return err
			}

			d.merge(dir)
		} else {
			d = dir
		}

		data, err := json.Marshal(&d)
		if err != nil {
			return err
		}

		return bucket.Put(key, data)
	})
}

func (s *dataStore) forEeachDirectory(fn func(d directory) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var d directory

			if err := json.Unmarshal(v, &d); err != nil {
				return err
			}

			if err := fn(d); err != nil {
				return err
			}
		}

		return nil
	})
}
