package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
)

const dir = "N:/Go"

func getTotalSize(dir string) (size int64, err error) {
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		size += fi.Size()

		return nil
	})

	return
}

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

	totalSize, err := getTotalSize(dir)
	if err != nil {
		log.Fatalf("unable to compute total size for %s. err=%v", dir, err)
	}

	log.Printf("total size for %s is %s", dir, humanize.Bytes(uint64(totalSize)))

	start := time.Now()

	var tb *tarball
	var fi os.FileInfo
	{
		tb = newTarball(dir, totalSize)
		if err := tb.process(); err != nil {
			log.Fatalf("unable to make tarball. err=%v", err)
		}

		fi, err = tb.Stat()
		if err != nil {
			log.Fatalf("unable to stat tarball. err=%v", err)
		}
	}

	{
		ch := make(chan float32)

		go func() {
			log.Printf("upload progress...")

			start := time.Now()

			for p := range ch {
				if time.Now().Sub(start) >= time.Second*1 {
					log.Printf("p=%f", p)
					start = time.Now()
				}
			}
		}()

		tr := newTrackedReader(tb, fi.Size(), ch)
		err = copier.CopyFromReader(tr, fi.Size(), fmt.Sprintf("%s_%s.tar.gz", filepath.Base(dir), time.Now().Format("2006-01-02")))
		if err != nil {
			log.Fatalf("unable to copy the tarball to the remote host. err=%v", err)
		}
		close(ch)
	}

	elapsed := time.Now().Sub(start)

	log.Printf("elapsed: %s", elapsed)

	if err := tb.Close(); err != nil {
		log.Fatalf("unable to close tarball. err=%v", err)
	}
}
