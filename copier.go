package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davidmz/go-pageant"

	"golang.org/x/crypto/ssh"
)

type scpRemoteCopier struct {
	logger  *logger
	scratch [0xFF]byte

	name   string
	params *sshParameters
	client *ssh.Client
}

func (c *scpRemoteCopier) String() string {
	return fmt.Sprintf("{name: %s, host: %s, port: %d, user: %s}",
		c.name, c.params.Host, c.params.Port, c.params.User,
	)
}

func newSCPRemoteCopier(logger *logger, name string, params *sshParameters) *scpRemoteCopier {
	return &scpRemoteCopier{
		logger: logger,
		name:   name,
		params: params,
	}
}

func (c *scpRemoteCopier) Connect() (err error) {
	var signers []ssh.Signer
	{
		if pageant.Available() {
			sshAgent := pageant.New()
			pageantSigners, err := sshAgent.Signers()
			if err != nil {
				return err
			}

			signers = pageantSigners
		}

		if c.params.PrivateKeyFile != "" {
			privateKeyBytes, err := ioutil.ReadFile(c.params.PrivateKeyFile)
			if err != nil {
				return err
			}

			signer, err := ssh.ParsePrivateKey(privateKeyBytes)
			if err != nil {
				return err
			}
			signers = append(signers, signer)
		}
	}

	clientConfig := ssh.ClientConfig{
		User: c.params.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signers...),
		},
	}

	addr := fmt.Sprintf("%s:%d", c.params.Host, c.params.Port)

	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return err
	}

	sshc, chans, reqs, err := ssh.NewClientConn(conn, addr, &clientConfig)
	if err != nil {
		return err
	}

	c.client = ssh.NewClient(sshc, chans, reqs)

	return nil
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

func (c *scpRemoteCopier) CopyFromReader(src io.Reader, size int64, path string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, err := session.StdinPipe()
		if err != nil {
			c.logger.Errorf(1, "unable to get stdin pipe. err=%v", err)
			return
		}

		r, err := session.StdoutPipe()
		if err != nil {
			c.logger.Errorf(1, "unable to get stdout pipe. err=%v", err)
			return
		}

		dir := filepath.Dir(path)
		if dir != "." {
			for _, d := range strings.Split(dir, string(os.PathSeparator)) {
				if err := c.addDirectory(w, r, d); err != nil {
					c.logger.Errorf(1, "unable to add directory. err=%v", err)
					return
				}
			}
		}

		fname := filepath.Base(path)
		_, err = fmt.Fprintf(w, "C0600 %d %s\n", size, fname)
		if err != nil {
			c.logger.Errorf(1, "unable to add file %s. err=%v", fname, err)
			return
		}

		{
			_, err = r.Read(c.scratch[0:0])
			if err != nil {
				c.logger.Errorf(1, "unable to read response byte. err=%v", err)
				return
			}
			if c.scratch[0] != 0 {
				err := c.readError(r, c.scratch[0])
				c.logger.Errorf(1, "response was %d. err=%v", c.scratch[0], err)
				return
			}
		}

		_, err = io.Copy(w, src)
		if err != nil {
			c.logger.Errorf(1, "unable to copy file data. err=%v", err)
			return
		}

		_, err = w.Write([]byte{0})
		if err != nil {
			c.logger.Errorf(1, "unable to write start data transfer byte. err=%v", err)
			return
		}

		{
			_, err = r.Read(c.scratch[0:0])
			if err != nil {
				c.logger.Errorf(1, "unable to read response byte. err=%v", err)
				return
			}
			if c.scratch[0] != 0 {
				err := c.readError(r, c.scratch[0])
				c.logger.Errorf(1, "response was %d. err=%v", c.scratch[0], err)
				return
			}
		}

		if err = w.Close(); err != nil {
			c.logger.Errorf(1, "unable to close stdin pipe. err=%v", err)
			return
		}
	}()

	return session.Run(fmt.Sprintf("/usr/bin/scp -tr %s", c.params.BackupsDir))
}

func (c *scpRemoteCopier) Close() error {
	return c.client.Close()
}
