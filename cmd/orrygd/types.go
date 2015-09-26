package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

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
