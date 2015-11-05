package main

import "time"

type configuration interface {
	ReadSCPCopiers() ([]scpCopierConf, error)
	ReadDirectories() ([]directory, error)
	ReadCheckFrequency() (time.Duration, error)
	ReadCleanupFrequency() (time.Duration, error)
	ReadDateFormat() (string, error)

	WriteSCPCopier(conf scpCopierConf) error
	WriteDirectory(d directory) error
	WriteCheckFrequency(d time.Duration) error
	WriteCleanupFrequency(d time.Duration) error
	WriteDateFormat(s string) error

	UpdateLastUpdated(d directory) error

	DumpConfig() ([]string, error)
}
