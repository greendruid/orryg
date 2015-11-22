package main

import "github.com/lxn/walk"

type directoriesModel struct {
	*walk.ListModelBase

	directories []directory
}

func newDirectoriesModel() (*directoriesModel, error) {
	dirs, err := conf.ReadDirectories()
	if err != nil {
		return nil, err
	}

	return &directoriesModel{
		ListModelBase: &walk.ListModelBase{},
		directories:   dirs,
	}, nil
}

func (d *directoriesModel) ItemCount() int {
	logger.Printf("called item count")
	return len(d.directories)
}

func (d *directoriesModel) Value(index int) interface{} {
	logger.Printf("called value for %d", index)
	return d.directories[index]
}
