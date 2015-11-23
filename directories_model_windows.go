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

func (m *directoriesModel) monitorDirChanged(ch chan directory) {
	for changedDir := range ch {
		for i, dd := range m.directories {
			if dd.Equal(changedDir) {
				m.directories[i] = changedDir
				m.PublishRowChanged(i)
			}
		}
	}
}

func (m *directoriesModel) RowCount() int {
	return len(m.directories)
}

const (
	colArchiveName = iota
	colOriginalPath
	colFrequency
	colMaxBackups
	colMaxBackupAge
	colLastUpdated
)

func (m *directoriesModel) Value(row, col int) interface{} {
	if row > len(m.directories) {
		return nil
	}

	dd := m.directories[row]

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
