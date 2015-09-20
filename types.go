package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type settings struct {
	CheckFrequency time.Duration
	DateFormat     string
}

func defaultSettings() settings {
	return settings{
		CheckFrequency: time.Minute * 1,
		DateFormat:     "20060201_030405",
	}
}

type scpCopierConf struct {
	Name   string
	Params sshParameters
}

type copierType uint8

const (
	unknownCopierType copierType = iota
	scpCopierType
)

func (t copierType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *copierType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	val, err := newCopierTypeFromString(s)
	if err != nil {
		return err
	}

	*t = val

	return nil
}

func (t copierType) String() string {
	switch t {
	case scpCopierType:
		return "scp"
	default:
		return "unknown"
	}
}

func newCopierTypeFromString(s string) (copierType, error) {
	switch strings.ToLower(s) {
	case "scp":
		return scpCopierType, nil
	default:
		return unknownCopierType, fmt.Errorf("unknown copier type %s", s)
	}
}

type directory struct {
	Frequency   time.Duration
	OrigPath    string
	ArchiveName string
	LastUpdated time.Time
}

func (d *directory) merge(id directory) {
	if id.Frequency > 0 {
		d.Frequency = id.Frequency
	}
	if id.OrigPath != "" {
		d.OrigPath = id.OrigPath
	}
	if id.ArchiveName != "" {
		d.ArchiveName = id.ArchiveName
	}
	if !id.LastUpdated.IsZero() {
		d.LastUpdated = id.LastUpdated
	}
}
