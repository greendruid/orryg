package main

import (
	"fmt"
	"time"
)

type sshParameters struct {
	User           string
	Host           string
	Port           int
	PrivateKeyFile string
	BackupsDir     string
}

type scpCopierConf struct {
	Name   string
	Params sshParameters
}

func (c scpCopierConf) String() string {
	const format = "{name: %s, user: %s, host: %s, port: %d, privateKeyFile: %s, backupsDir: %s}"
	return fmt.Sprintf(format,
		c.Name, c.Params.User, c.Params.Host, c.Params.Port,
		c.Params.PrivateKeyFile, c.Params.BackupsDir,
	)
}

type directory struct {
	Frequency    time.Duration
	OriginalPath string
	ArchiveName  string
	MaxBackups   int
	MaxBackupAge time.Duration
	LastUpdated  time.Time
}

func (d *directory) Equal(other directory) bool {
	return d.ArchiveName == other.ArchiveName && d.OriginalPath == other.OriginalPath
}

func (d *directory) UniqueID() string {
	return d.ArchiveName + d.OriginalPath
}

func (d directory) String() string {
	const format = "{frequency: %s, originalPath: %s, archiveName: %s, lastUpdated: %s}"
	return fmt.Sprintf(format,
		d.Frequency, d.OriginalPath,
		d.ArchiveName, d.LastUpdated,
	)
}
