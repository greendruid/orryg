package main

import "github.com/lxn/walk"

type copiersModel struct {
	*walk.TableModelBase

	copiers []scpCopierConf
}

func newCopiersModel() (*copiersModel, error) {
	copiers, err := conf.ReadSCPCopiers()
	if err != nil {
		return nil, err
	}

	return &copiersModel{
		TableModelBase: &walk.TableModelBase{},
		copiers:        copiers,
	}, nil
}

func (d *copiersModel) RowCount() int {
	return len(d.copiers)
}

const (
	colName = iota
	colUser
	colHost
	colPort
	colPrivateKeyFile
	colBackupsDir
)

func (d *copiersModel) Value(row, col int) interface{} {
	if row > len(d.copiers) {
		return nil
	}

	c := d.copiers[row]

	switch col {
	case colName:
		return c.Name
	case colUser:
		return c.Params.User
	case colHost:
		return c.Params.Host
	case colPort:
		return c.Params.Port
	case colPrivateKeyFile:
		return c.Params.PrivateKeyFile
	case colBackupsDir:
		return c.Params.BackupsDir
	default:
		return nil
	}
}
