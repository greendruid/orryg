package main

import (
	"encoding/json"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/vrischmann/orryg"
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

func (s *dataStore) getSettings() (se orryg.Settings, err error) {
	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(settingsBucket)

		data := bucket.Get(settingsBucket)
		if data == nil {
			se = orryg.DefaultSettings()

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

func (s *dataStore) removeCopier(name string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(copiersBucket)

		cursor := bucket.Cursor()
		for k, _ := cursor.First(); k != nil; k, _ = cursor.Next() {
			if string(k[:len(k)-1]) == name {
				if err := cursor.Delete(); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *dataStore) getAllSCPCopierConfs() (res []orryg.SCPCopierConf, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(copiersBucket)

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			typ := orryg.CopierType(k[len(k)-1])
			if typ != orryg.SCPCopierType {
				continue
			}

			var params orryg.SSHParameters
			if err := json.Unmarshal(v, &params); err != nil {
				return err
			}

			res = append(res, orryg.SCPCopierConf{
				Name:   string(k[:len(k)-1]),
				Params: params,
			})
		}

		return nil
	})

	return
}

func (s *dataStore) mergeSCPCopierConf(c orryg.SCPCopierConf) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(copiersBucket)

		key := append([]byte(c.Name), byte(orryg.SCPCopierType))

		var params orryg.SSHParameters
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &params); err != nil {
				return err
			}

			params.Merge(c.Params)
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

func (s *dataStore) mergeDirectory(dir orryg.Directory) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		key := []byte(dir.OriginalPath)

		var d orryg.Directory
		if data := bucket.Get(key); data != nil {
			if err := json.Unmarshal(data, &d); err != nil {
				return err
			}

			d.Merge(dir)
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

func (s *dataStore) forEeachDirectory(fn func(d orryg.Directory) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(directoriesBucket)

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var d orryg.Directory

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
