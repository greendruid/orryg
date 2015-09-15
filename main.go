package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"
)

const dir = "N:/Projects/json"

func main() {
	copier := newSSHRemoteCopier(sshParameters{
		user:           "sphax",
		host:           "192.168.1.34",
		port:           22,
		privateKeyFile: "N:/backup_dev.rsa",
		backupsDir:     "/home/sphax/backups",
	})
	defer copier.Close()

	if err := copier.Connect(); err != nil {
		log.Fatalln(err)
	}

	start := time.Now()

	tb := newTarball(dir)
	if err := tb.process(); err != nil {
		log.Fatalf("unable to make tarball. err=%v", err)
	}

	fi, err := tb.Stat()
	if err != nil {
		log.Fatalf("unable to stat tarball. err=%v", err)
	}

	err = copier.CopyFromReader(tb, fi.Size(), fmt.Sprintf("%s_%s.tar.gz", filepath.Base(dir), time.Now().Format("2006-01-02")))
	if err != nil {
		log.Fatalf("unable to copy the tarball to the remote host. err=%v", err)
	}

	elapsed := time.Now().Sub(start)

	log.Printf("elapsed: %s", elapsed)

	if err := tb.Close(); err != nil {
		log.Fatalf("unable to close tarball. err=%v", err)
	}
}
