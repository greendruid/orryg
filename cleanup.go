package main

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"
)

type cleaner struct {
	backupsDir string
	client     *sshClient
}

func newCleaner(params *sshParameters) *cleaner {
	return &cleaner{
		backupsDir: params.BackupsDir,
		client:     newSSHClient(params),
	}
}

func (c *cleaner) cleanAllExpiredBackups(id directory, dateFormat string) error {
	filenames, err := c.getExpiredBackups(id, dateFormat)
	if err != nil {
		return fmt.Errorf("unable to get expired backups for %v. err=%v", id, err)
	}

	for _, filename := range filenames {
		logger.Printf("cleaning %s", filename)

		if err := c.clean(filename); err != nil {
			logger.Printf("unable to clean %s. err=%v", filename, err)
		}
	}

	return nil
}

func (c *cleaner) clean(filename string) error {
	session := c.client.getValidSession()
	defer session.Close()

	// NOTE(vincent): do NOT use filepath.Join because it produces a path with \ since we're on windows.
	fullPath := c.backupsDir + "/" + filename

	data, err := session.CombinedOutput(fmt.Sprintf("/bin/rm %s", fullPath))
	if err != nil {
		return fmt.Errorf("unable to clean file %s. output=%s err=%v", fullPath, string(data), err)
	}

	return nil
}

func (c *cleaner) getExpiredBackups(d directory, dateFormat string) (res []string, err error) {
	session := c.client.getValidSession()
	defer session.Close()

	data, err := session.CombinedOutput(fmt.Sprintf("/bin/ls %s", c.backupsDir))
	if err != nil {
		return nil, err
	}

	var els elements

	// Create the list of potentially valid elements
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		filename := line

		// If not for this archive don't do anything
		if !strings.HasPrefix(line, d.ArchiveName) {
			continue
		}

		var tstr string
		{
			line = line[len(d.ArchiveName)+1:]

			pos := strings.Index(line, ".")
			if pos < 0 {
				continue
			}

			tstr = line[:pos]
		}

		t, err := time.Parse(dateFormat, tstr)
		if err != nil {
			logger.Printf("unable to parse date '%s'. err=%v", tstr, err)
			continue
		}

		els = append(els, el{filename, t})
	}

	// If we configured a TTL
	if d.MaxBackupAge > 0 {
		t := time.Now().UTC().Add(-d.MaxBackupAge)
		// return the position of the last element which has to be expired
		// takes care of sorting the elements.
		lastBefore := els.lastBefore(t)

		if lastBefore >= 0 {
			for _, v := range els[:lastBefore] {
				res = append(res, v.name)
			}

			return
		}
	}

	// If we configured a maximum number of backups
	if d.MaxBackups > 0 {
		sort.Sort(els)

		// Make sure there's more than the max
		if len(els) < d.MaxBackups {
			return
		}

		pos := len(els) - d.MaxBackups
		for _, v := range els[:pos] {
			res = append(res, v.name)
		}
	}

	return
}

func (c *cleaner) Close() error {
	return c.client.Close()
}

type el struct {
	name string
	t    time.Time
}

type elements []el

func (e elements) Len() int           { return len(e) }
func (e elements) Less(i, j int) bool { return e[i].t.Before(e[j].t) }
func (e elements) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

func (e elements) lastBefore(t time.Time) int {
	sort.Sort(e)

	m := -1
	for i, v := range e {
		if v.t.Before(t) {
			m = i
		}
	}

	return m
}
