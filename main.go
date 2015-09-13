package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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

	tf, err := ioutil.TempFile("", "orryg_tar")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		name := tf.Name()

		if err := tf.Close(); err != nil {
			log.Fatalln(err)
		}

		if err := os.Remove(name); err != nil {
			log.Fatalln(err)
		}
	}()

	aw := tar.NewWriter(tf)

	if err := aw.WriteHeader(&tar.Header{
		Name:     "json/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}); err != nil {
		log.Fatalln(err)
	}

	dir := "N:/Projects/json"
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		hdr.Name = filepath.Join("json", relPath)
		hdr.Name = strings.Replace(hdr.Name, string(os.PathSeparator), "/", -1)

		if err := aw.WriteHeader(hdr); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(aw, f)
		return err
	})
	if err != nil {
		log.Fatalln(err)
	}
	if err := aw.Close(); err != nil {
		log.Fatalln(err)
	}

	gzf, err := ioutil.TempFile("", "orryg_targz")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		name := gzf.Name()

		if err := gzf.Close(); err != nil {
			log.Fatalln(err)
		}

		if err := os.Remove(name); err != nil {
			log.Fatalln(err)
		}
	}()

	tf.Seek(0, os.SEEK_SET)

	gzw := gzip.NewWriter(gzf)

	_, err = io.Copy(gzw, tf)
	if err != nil {
		log.Fatalln(err)
	}

	if err := gzw.Close(); err != nil {
		log.Fatalln(err)
	}

	gzf.Seek(0, os.SEEK_SET)
	fi, err := gzf.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	err = copier.CopyFromReader(gzf, fi.Size(), fmt.Sprintf("%s_%s.tar.gz", filepath.Base(dir), time.Now().Format("2006-01-02")))
	log.Printf("%T %v", err, err)
}
