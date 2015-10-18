package main

import (
	"time"

	"github.com/vrischmann/jsonutil"
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
