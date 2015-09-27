package orryg

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Settings struct {
	CheckFrequency time.Duration `json:"checkFrequency"`
	DateFormat     string        `json:"dateFormat"`
}

func DefaultSettings() Settings {
	return Settings{
		CheckFrequency: time.Second * 1,
		DateFormat:     "20060201_030405",
	}
}

type CopierType uint8

const (
	UnknownCopierType CopierType = iota
	SCPCopierType
)

func (t CopierType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *CopierType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	val, err := NewCopierTypeFromString(s)
	if err != nil {
		return err
	}

	*t = val

	return nil
}

func (t CopierType) String() string {
	switch t {
	case SCPCopierType:
		return "scp"
	default:
		return "unknown"
	}
}

func NewCopierTypeFromString(s string) (CopierType, error) {
	switch strings.ToLower(s) {
	case "scp":
		return SCPCopierType, nil
	default:
		return UnknownCopierType, fmt.Errorf("unknown copier type %s", s)
	}
}

type SSHParameters struct {
	User           string `json:"user"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	PrivateKeyFile string `json:"privateKeyFile"`
	BackupsDir     string `json:"backupsDir"`
}

func (p *SSHParameters) Merge(params SSHParameters) {
	if params.User != "" {
		p.User = params.User
	}
	if params.Host != "" {
		p.Host = params.Host
	}
	if params.Port > 0 {
		p.Port = params.Port
	}
	if params.PrivateKeyFile != "" {
		p.PrivateKeyFile = params.PrivateKeyFile
	}
	if params.BackupsDir != "" {
		p.BackupsDir = params.BackupsDir
	}
}

type SCPCopierConf struct {
	Name   string        `json:"name"`
	Params SSHParameters `json:"params"`
}

type CopierConf struct {
	Type CopierType  `json:"type"`
	Conf interface{} `json:"conf"`
}

type CopiersConf []CopierConf

type UCopierConf struct {
	Type CopierType      `json:"type"`
	Conf json.RawMessage `json:"conf"`
}

type Directory struct {
	Frequency    time.Duration `json:"frequency"`
	OriginalPath string        `json:"originalPath"`
	ArchiveName  string        `json:"archiveName"`
	LastUpdated  time.Time     `json:"lastUpdated"`
}

type Directories []Directory

func (d *Directory) Merge(id Directory) {
	if id.Frequency > 0 {
		d.Frequency = id.Frequency
	}
	if id.OriginalPath != "" {
		d.OriginalPath = id.OriginalPath
	}
	if id.ArchiveName != "" {
		d.ArchiveName = id.ArchiveName
	}
	if !id.LastUpdated.IsZero() {
		d.LastUpdated = id.LastUpdated
	}
}
