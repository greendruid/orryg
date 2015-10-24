package main

import (
	"fmt"
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
	MaxBackups   int               `json:"maxBackups"`
	MaxBackupAge jsonutil.Duration `json:"maxBackupAge"`
	LastUpdated  time.Time         `json:"lastUpdated,omitempty"`
}

func (d directory) String() string {
	return fmt.Sprintf("{frequency: %s, originalPath: %s, archiveName: %s, lastUpdated: %s}",
		d.Frequency, d.OriginalPath, d.ArchiveName, d.LastUpdated,
	)
}

type config struct {
	SCPCopiers       []scpCopierConf   `json:"scpCopiers"`
	Directories      []*directory      `json:"directories"`
	CheckFrequency   jsonutil.Duration `json:"checkFrequency"`
	CleanupFrequency jsonutil.Duration `json:"cleanupFrequency"`
	DateFormat       string            `json:"dateFormat"`
}

func (c *config) update(id *directory) {
	for _, v := range c.Directories {
		if v.ArchiveName == id.ArchiveName {
			v.LastUpdated = time.Now()
		}
	}
}
