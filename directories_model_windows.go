package main

import (
	"time"

	"github.com/lxn/walk"
)

type directoriesModel struct {
	*walk.TableModelBase

	directories []directory
}

func newDirectoriesModel() (*directoriesModel, error) {
	dirs, err := conf.ReadDirectories()
	if err != nil {
		return nil, err
	}

	return &directoriesModel{
		TableModelBase: &walk.TableModelBase{},
		directories:    dirs,
	}, nil
}

func (d *directoriesModel) RowCount() int {
	return len(d.directories)
}

const (
	colArchiveName = iota
	colOriginalPath
	colFrequency
	colMaxBackups
	colMaxBackupAge
	colLastUpdated
)

func (d *directoriesModel) Value(row, col int) interface{} {
	if row > len(d.directories) {
		return nil
	}

	dd := d.directories[row]

	switch col {
	case colArchiveName:
		return dd.ArchiveName
	case colOriginalPath:
		return dd.OriginalPath
	case colFrequency:
		return dd.Frequency
	case colMaxBackups:
		return dd.MaxBackups
	case colMaxBackupAge:
		return dd.MaxBackupAge
	case colLastUpdated:
		return dd.LastUpdated.Format(time.RFC3339)
	default:
		return nil
	}
}
