// +build windows
package main

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/registry"
)

type windowsConfiguration struct {
}

func newWindowsConfiguration() configuration {
	return &windowsConfiguration{}
}

func (c *windowsConfiguration) ReadSCPCopiers() (res []scpCopierConf, err error) {
	names, err := c.readSCPCopiersNames()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		conf, err := c.readSCPCopier(name)
		if err != nil {
			return nil, err
		}

		res = append(res, conf)
	}

	return
}

func withKey(keyName string, fn func(key registry.Key) error) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `SOFTWARE\Orryg`+keyName, registry.ALL_ACCESS)
	if err != nil {
		return err
	}
	defer key.Close()

	return fn(key)
}

func (c *windowsConfiguration) readSCPCopier(name string) (conf scpCopierConf, err error) {
	err = withKey(`\SCPCopiers\`+name, func(key registry.Key) error {
		var der delayedErrorRegKey

		copierName := der.musts(key.GetStringValue("Name"))
		user := der.musts(key.GetStringValue("User"))
		host := der.musts(key.GetStringValue("Host"))
		port := der.musti(key.GetIntegerValue("Port"))
		privateKeyFile := der.musts(key.GetStringValue("PrivateKeyFile"))
		backupsDir := der.musts(key.GetStringValue("BackupsDir"))

		conf = scpCopierConf{
			Name: copierName,
			Params: sshParameters{
				User:           user,
				Host:           host,
				Port:           int(port),
				PrivateKeyFile: privateKeyFile,
				BackupsDir:     backupsDir,
			},
		}

		return der.err
	})
	return
}

func (c *windowsConfiguration) readSCPCopiersNames() (vals []string, err error) {
	err = withKey("", func(key registry.Key) error {
		var r error
		vals, _, r = key.GetStringsValue("SCPCopiersNames")
		if r == registry.ErrNotExist {
			return nil
		}
		return r
	})
	return
}

func (c *windowsConfiguration) ReadDirectories() (res []directory, err error) {
	names, err := c.readDirectoriesNames()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		d, err := c.readDirectory(name)
		if err != nil {
			return nil, err
		}

		res = append(res, d)
	}

	return
}

func (c *windowsConfiguration) readDirectory(name string) (d directory, err error) {
	err = withKey(`\Directories\`+name, func(key registry.Key) error {
		var der delayedErrorRegKey

		frequency := der.musti(key.GetIntegerValue("Frequency"))
		originalPath := der.musts(key.GetStringValue("OriginalPath"))
		archiveName := der.musts(key.GetStringValue("ArchiveName"))
		maxBackups := der.musti(key.GetIntegerValue("MaxBackups"))
		maxBackupAge := der.musti(key.GetIntegerValue("MaxBackupAge"))
		lastUpdated := der.musts(key.GetStringValue("LastUpdated"))

		if der.err != nil {
			return der.err
		}

		lastUpdatedTime, r := time.Parse(time.RFC3339, lastUpdated)
		if r != nil {
			lastUpdatedTime = time.Time{}
		}

		d = directory{
			Frequency:    time.Duration(frequency),
			OriginalPath: originalPath,
			ArchiveName:  archiveName,
			MaxBackups:   int(maxBackups),
			MaxBackupAge: time.Duration(maxBackupAge),
			LastUpdated:  lastUpdatedTime,
		}

		return nil
	})
	return
}

func (c *windowsConfiguration) readDirectoriesNames() (vals []string, err error) {
	err = withKey("", func(key registry.Key) error {
		var r error
		vals, _, r = key.GetStringsValue("DirectoriesNames")
		if r == registry.ErrNotExist {
			return nil
		}
		return r
	})
	return
}

func (c *windowsConfiguration) ReadCheckFrequency() (d time.Duration, err error) {
	err = withKey("", func(key registry.Key) error {
		var r error
		val, _, r := key.GetIntegerValue("CheckFrequency")
		if r == registry.ErrNotExist {
			r = nil
		}
		d = time.Duration(int64(val))
		return r
	})
	return
}

func (c *windowsConfiguration) ReadCleanupFrequency() (d time.Duration, err error) {
	err = withKey("", func(key registry.Key) error {
		var r error
		val, _, r := key.GetIntegerValue("CleanupFrequency")
		if r == registry.ErrNotExist {
			r = nil
		}
		d = time.Duration(int64(val))
		return r
	})
	return
}

func (c *windowsConfiguration) ReadDateFormat() (f string, err error) {
	err = withKey("", func(key registry.Key) error {
		var r error
		f, _, r = key.GetStringValue("DateFormat")
		if r == registry.ErrNotExist {
			r = nil
		}
		return r
	})
	return
}

func appendAndUniq(sl []string, s string) (res []string) {
	m := make(map[string]struct{})
	for _, el := range sl {
		m[el] = struct{}{}
	}
	m[s] = struct{}{}

	for k, _ := range m {
		res = append(res, k)
	}

	return
}

func (c *windowsConfiguration) WriteSCPCopier(conf scpCopierConf) error {
	err := withKey("", func(key registry.Key) error {
		vals, _, r := key.GetStringsValue("SCPCopiersNames")
		if r != nil && r != registry.ErrNotExist {
			return r
		}

		return key.SetStringsValue("SCPCopiersNames", appendAndUniq(vals, conf.Name))
	})
	if err != nil {
		return err
	}

	return withKey(`\SCPCopiers\`+conf.Name, func(key registry.Key) error {
		var der delayedErrorRegKey

		der.must(key.SetStringValue("Name", conf.Name))
		der.must(key.SetStringValue("User", conf.Params.User))
		der.must(key.SetStringValue("Host", conf.Params.Host))
		der.must(key.SetDWordValue("Port", uint32(conf.Params.Port)))
		der.must(key.SetStringValue("PrivateKeyFile", conf.Params.PrivateKeyFile))
		der.must(key.SetStringValue("BackupsDir", conf.Params.BackupsDir))

		return der.err
	})
}

func (c *windowsConfiguration) WriteDirectory(d directory) error {
	err := withKey("", func(key registry.Key) error {
		vals, _, r := key.GetStringsValue("DirectoriesNames")
		if r != nil && r != registry.ErrNotExist {
			return r
		}

		return key.SetStringsValue("DirectoriesNames", appendAndUniq(vals, d.ArchiveName))
	})
	if err != nil {
		return err
	}

	return withKey(`\Directories\`+d.ArchiveName, func(key registry.Key) error {
		var der delayedErrorRegKey

		der.must(key.SetQWordValue("Frequency", uint64(d.Frequency)))
		der.must(key.SetStringValue("OriginalPath", d.OriginalPath))
		der.must(key.SetStringValue("ArchiveName", d.ArchiveName))
		der.must(key.SetDWordValue("MaxBackups", uint32(d.MaxBackups)))
		der.must(key.SetQWordValue("MaxBackupAge", uint64(d.MaxBackupAge)))
		der.must(key.SetStringValue("LastUpdated", d.LastUpdated.Format(time.RFC3339)))

		return der.err
	})
}

func (c *windowsConfiguration) WriteCheckFrequency(d time.Duration) error {
	return withKey("", func(key registry.Key) error {
		return key.SetQWordValue("CheckFrequency", uint64(d))
	})
}

func (c *windowsConfiguration) WriteCleanupFrequency(d time.Duration) error {
	return withKey("", func(key registry.Key) error {
		return key.SetQWordValue("CleanupFrequency", uint64(d))
	})
}

func (c *windowsConfiguration) WriteDateFormat(s string) error {
	return withKey("", func(key registry.Key) error {
		return key.SetStringValue("DateFormat", s)
	})
}

func (c *windowsConfiguration) UpdateLastUpdated(d directory) error {
	return withKey("", func(key registry.Key) error {
		return key.SetStringValue("LastUpdated", d.LastUpdated.Format(time.RFC3339))
	})
}

func (c *windowsConfiguration) DumpConfig() (res []string, err error) {
	// {
	// 	copiers, err := c.ReadSCPCopiers()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	//
	// 	if len(copiers) > 0 {
	// 		res = append(res, "copiers:")
	// 		for _, conf := range copiers {
	// 			res = append(res, conf.String())
	// 		}
	// 	}
	// }
	//
	{
		directories, err := c.ReadDirectories()
		if err != nil {
			return nil, err
		}

		if len(directories) > 0 {
			res = append(res, "directories:")
			for _, d := range directories {
				res = append(res, d.String())
			}
		}
	}

	{
		freq, err := c.ReadCheckFrequency()
		if err != nil {
			return nil, err
		}

		res = append(res, fmt.Sprintf("check frequency: %s", freq.String()))
	}

	{
		freq, err := c.ReadCleanupFrequency()
		if err != nil {
			return nil, err
		}

		res = append(res, fmt.Sprintf("cleanup frequency: %s", freq.String()))
	}

	{
		format, err := c.ReadDateFormat()
		if err != nil {
			return nil, err
		}

		res = append(res, fmt.Sprintf("date format: %s", format))
	}

	return
}

type delayedErrorRegKey struct {
	err error
}

func (d *delayedErrorRegKey) must(err error) {
	if d.err != nil {
		return
	}
	d.err = err
}

func (d *delayedErrorRegKey) musts(s string, vt uint32, err error) string {
	d.err = err
	if d.err != nil {
		return ""
	}

	return s
}

func (d *delayedErrorRegKey) musti(s uint64, vt uint32, err error) int64 {
	d.err = err
	if d.err != nil {
		return 0
	}

	return int64(s)
}
