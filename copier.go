package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh"
)

type remoteCopier interface {
	CopyFromReader(src io.Reader, size int64, path string) error
	Connect() error
	Close() error
}

type dummyRemoteCopier struct{}

func (f dummyRemoteCopier) CopyFromReader(src io.Reader, size int64, path string) error {
	fmt.Printf("path: %s size: %d\n", path, size)
	n, err := io.Copy(os.Stdout, src)
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("only read %d but expected %d", n, size)
	}

	return nil
}
func (f dummyRemoteCopier) Connect() error { return nil }
func (f dummyRemoteCopier) Close() error   { return nil }

type sshParameters struct {
	user           string
	host           string
	port           int
	privateKeyFile string
	backupsDir     string
}

type sshRemoteCopier struct {
	params sshParameters

	client *ssh.Client
}

func newSSHRemoteCopier(params sshParameters) remoteCopier {
	return &sshRemoteCopier{params: params}
}

func (c *sshRemoteCopier) Connect() error {
	privateKeyBytes, err := ioutil.ReadFile(c.params.privateKeyFile)
	if err != nil {
		return err
	}

	signer, err := ssh.ParsePrivateKey(privateKeyBytes)
	if err != nil {
		return err
	}

	clientConfig := ssh.ClientConfig{
		User: c.params.user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	c.client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.params.host, c.params.port), &clientConfig)
	if err != nil {
		return err
	}

	return err
}

var (
	scratch [0xFF]byte
)

func readError(r io.Reader, code byte) error {
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

func addDirectory(w io.WriteCloser, r io.Reader, name string) (err error) {
	_, err = fmt.Fprintf(w, "D0755 0 %s\n", name)
	log.Printf("D0755 0 %s", name)
	if err != nil {
		return err
	}

	_, err = r.Read(scratch[0:0])
	if err != nil {
		return err
	}
	if scratch[0] != 0 {
		return readError(r, scratch[0])
	}

	return nil
}

func (c *sshRemoteCopier) CopyFromReader(src io.Reader, size int64, path string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, err := session.StdinPipe()
		if err != nil {
			log.Printf("unable to get stdin pipe. err=%v", err)
			return
		}

		r, err := session.StdoutPipe()
		if err != nil {
			log.Printf("unable to get stdout pipe. err=%v", err)
			return
		}

		dir := filepath.Dir(path)
		if dir != "." {
			for _, d := range strings.Split(dir, string(os.PathSeparator)) {
				if err := addDirectory(w, r, d); err != nil {
					log.Printf("unable to add directory. err=%v", err)
					return
				}
			}
		}

		fname := filepath.Base(path)
		_, err = fmt.Fprintf(w, "C0600 %d %s\n", size, fname)
		log.Printf("C0600 %d %s", size, fname)
		if err != nil {
			log.Printf("unable to add file %s. err=%v", fname, err)
			return
		}

		{
			_, err = r.Read(scratch[0:0])
			if err != nil {
				log.Printf("unable to read response byte. err=%v", err)
				return
			}
			if scratch[0] != 0 {
				err := readError(r, scratch[0])
				log.Printf("response was %d. err=%v", scratch[0], err)
				return
			}
		}

		n, err := io.Copy(w, src)
		if err != nil {
			log.Printf("unable to copy file data. err=%v", err)
			return
		}
		log.Printf("written %d", n)

		_, err = w.Write([]byte{0})
		if err != nil {
			log.Printf("unable to write start data transfer byte. err=%v", err)
			return
		}

		{
			_, err = r.Read(scratch[0:0])
			if err != nil {
				log.Printf("unable to read response byte. err=%v", err)
				return
			}
			if scratch[0] != 0 {
				err := readError(r, scratch[0])
				log.Printf("response was %d. err=%v", scratch[0], err)
				return
			}
		}

		if err = w.Close(); err != nil {
			log.Printf("unable to close stdin pipe. err=%v", err)
			return
		}
	}()

	return session.Run(fmt.Sprintf("/usr/bin/scp -tr %s", c.params.backupsDir))
}

func (c *sshRemoteCopier) Close() error {
	return c.client.Close()
}
