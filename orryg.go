package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/vrischmann/jsonutil"
	"github.com/vrischmann/userdir"
)

type sshParameters struct {
	User           string `json:"user"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	PrivateKeyFile string `json:"privateKeyFile"`
	BackupsDir     string `json:"backupsDir"`
}

type scpCopierConf struct {
	Name   string        `json:"name"`
	Params sshParameters `json:"params"`
}

type directory struct {
	Frequency    jsonutil.Duration `json:"frequency"`
	OriginalPath string            `json:"originalPath"`
	ArchiveName  string            `json:"archiveName"`
	LastUpdated  time.Time         `json:"lastUpdated,omitempty"`
}

type config struct {
	SCPCopiers     []scpCopierConf   `json:"scpCopiers"`
	Directories    []*directory      `json:"directories"`
	CheckFrequency jsonutil.Duration `json:"checkFrequency"`
	DateFormat     string            `json:"dateFormat"`
}

// func (c *config) setLastUpdated(d *directory, t time.Time) {
// 	for _, v := range c.Directories {
// 		if v.OriginalPath == d.OriginalPath {
// 			v.LastUpdated = t
// 		}
// 	}
// }

var configPath = filepath.Join(userdir.GetConfigHome(), "orryg", "config.json")

func readConfig() error {
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)

	return dec.Decode(&conf)
}

func writeConfig() error {
	f, err := os.OpenFile(configPath, os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(&conf, "", "    ")
	if err != nil {
		return err
	}

	_, err = f.Write(data)

	return err
}
