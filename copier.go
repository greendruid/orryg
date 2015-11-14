package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type scpRemoteCopier struct {
	scratch [0xFF]byte

	name       string
	backupsDir string
	client     *sshClient
}

func (c *scpRemoteCopier) String() string {
	return fmt.Sprintf("{name: %s, client: %s}", c.name, c.client)
}

func newSCPRemoteCopier(name string, params *sshParameters) *scpRemoteCopier {
	return &scpRemoteCopier{
		name:       name,
		backupsDir: params.BackupsDir,
		client:     newSSHClient(params),
	}
}

func (c *scpRemoteCopier) readError(r io.Reader, code byte) error {
	var text string
	_, err := fmt.Fscanln(r, &text)
	if err != nil {
		return err
	}

	if text == "" {
		return fmt.Errorf("unknown error. code=%d", code)
	}

	return fmt.Errorf("%s. code=%d", text, code)
}

func (c *scpRemoteCopier) addDirectory(w io.WriteCloser, r io.Reader, name string) (err error) {
	_, err = fmt.Fprintf(w, "D0755 0 %s\n", name)
	if err != nil {
		return err
	}

	_, err = r.Read(c.scratch[0:0])
	if err != nil {
		return err
	}
	if c.scratch[0] != 0 {
		return c.readError(r, c.scratch[0])
	}

	return nil
}

func (c *scpRemoteCopier) CopyFromReader(src io.Reader, size int64, path string) (err error) {
	session := c.client.getValidSession()
	defer session.Close()

	go func() {
		w, err := session.StdinPipe()
		if err != nil {
			logger.Printf("unable to get stdin pipe. err=%v", err)
			return
		}

		r, err := session.StdoutPipe()
		if err != nil {
			logger.Printf("unable to get stdout pipe. err=%v", err)
			return
		}

		dir := filepath.Dir(path)
		if dir != "." {
			for _, d := range strings.Split(dir, string(os.PathSeparator)) {
				if err := c.addDirectory(w, r, d); err != nil {
					logger.Printf("unable to add directory. err=%v", err)
					return
				}
			}
		}

		fname := filepath.Base(path)
		_, err = fmt.Fprintf(w, "C0600 %d %s\n", size, fname)
		if err != nil {
			logger.Printf("unable to add file %s. err=%v", fname, err)
			return
		}

		{
			_, err = r.Read(c.scratch[0:0])
			if err != nil {
				logger.Printf("unable to read response byte. err=%v", err)
				return
			}
			if c.scratch[0] != 0 {
				err := c.readError(r, c.scratch[0])
				logger.Printf("response was %d. err=%v", c.scratch[0], err)
				return
			}
		}

		_, err = io.Copy(w, src)
		if err != nil {
			logger.Printf("unable to copy file data. err=%v", err)
			return
		}

		_, err = w.Write([]byte{0})
		if err != nil {
			logger.Printf("unable to write start data transfer byte. err=%v", err)
			return
		}

		{
			_, err = r.Read(c.scratch[0:0])
			if err != nil {
				logger.Printf("unable to read response byte. err=%v", err)
				return
			}
			if c.scratch[0] != 0 {
				err := c.readError(r, c.scratch[0])
				logger.Printf("response was %d. err=%v", c.scratch[0], err)
				return
			}
		}

		if err = w.Close(); err != nil {
			logger.Printf("unable to close stdin pipe. err=%v", err)
			return
		}
	}()

	return session.Run(fmt.Sprintf("/usr/bin/scp -tr %s", c.backupsDir))
}

func (c *scpRemoteCopier) Close() error {
	return c.client.Close()
}
