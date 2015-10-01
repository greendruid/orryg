package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type tarball struct {
	d *directory

	tf  *os.File
	aw  *tar.Writer
	gzf *os.File
	ef  *os.File
	fi  os.FileInfo

	err    error
	copied int64
}

func newTarball(d *directory) *tarball {
	return &tarball{d: d}
}

func (t *tarball) process() error {
	t.makeTar()
	t.populateTar()
	t.makeTarball()
	t.encryptTarball()
	t.stat()

	return t.err
}

func (t *tarball) makeTar() {
	if t.err != nil {
		return
	}

	t.tf, t.err = ioutil.TempFile("", "orryg_tar")
	if t.err != nil {
		return
	}

	t.aw = tar.NewWriter(t.tf)

	// This creates a single root directory in the tarball.
	// Prevents polluting the cwd when untar-ing the archive.
	t.err = t.aw.WriteHeader(&tar.Header{
		Name:     t.d.ArchiveName,
		Mode:     0755,
		Typeflag: tar.TypeDir,
	})
}

func (t *tarball) populateTar() {
	t.err = filepath.Walk(t.d.OriginalPath, func(path string, info os.FileInfo, err error) error {
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

		relPath, err := filepath.Rel(t.d.OriginalPath, path)
		if err != nil {
			return err
		}

		hdr.Name = filepath.Join(t.d.ArchiveName, relPath)
		hdr.Name = strings.Replace(hdr.Name, string(os.PathSeparator), "/", -1)

		if err := t.aw.WriteHeader(hdr); err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(t.aw, f)

		return err
	})

	if t.err != nil {
		return
	}

	t.err = t.aw.Close()
}

func (t *tarball) makeTarball() {
	t.gzf, t.err = ioutil.TempFile("", "orryg_targz")
	if t.err != nil {
		return
	}

	t.tf.Seek(0, os.SEEK_SET)

	gzw := gzip.NewWriter(t.gzf)

	_, t.err = io.Copy(gzw, t.tf)
	if t.err != nil {
		return
	}

	t.err = gzw.Close()
	if t.err != nil {
		return
	}

	t.err = t.tf.Close()
	if t.err != nil {
		return
	}

	t.err = os.Remove(t.tf.Name())
	if t.err != nil {
		return
	}
	t.tf = nil

	// Don't expect the caller to know it needs to seek back
	_, t.err = t.gzf.Seek(0, os.SEEK_SET)
}

func (t *tarball) encryptTarball() {
	// t.ef, t.err = ioutil.TempFile("", "orryg_gpg")
	// if t.err != nil {
	// 	return
	// }

	t.ef = t.gzf
}

func (t *tarball) stat() {
	t.fi, t.err = t.ef.Stat()
}

func (t *tarball) Read(p []byte) (int, error) {
	return t.ef.Read(p)
}

func (t *tarball) Close() error {
	if t.err != nil {
		return t.err
	}

	if err := t.ef.Close(); err != nil {
		return err
	}

	return os.Remove(t.ef.Name())
}
