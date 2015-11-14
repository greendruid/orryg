package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/davidmz/go-pageant"

	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	params *sshParameters
	client *ssh.Client
}

func (c sshClient) String() string {
	return fmt.Sprintf("{host: %s, port: %d, user: %s}", c.params.Host, c.params.Port, c.params.User)
}

func newSSHClient(params *sshParameters) *sshClient {
	return &sshClient{params: params}
}

func (c *sshClient) connect() (err error) {
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

func (c *sshClient) Close() error {
	return c.client.Close()
}

func (c *sshClient) getValidSession() (session *ssh.Session) {
	bo := backoff{
		duration:    time.Second,
		maxDuration: time.Second * 30,
	}

	var err error

	for {
		session, err = c.client.NewSession()
		if err == nil {
			break
		}

		logger.Printf("unable to create SSH session. err=%v", err)
		logger.Printf("trying to reconnect SSH client.")

		if err = c.connect(); err != nil {
			logger.Printf("unable to reconnect SSH client, retrying later. err=%v", err)
			bo.sleep()
		}
	}

	return
}
