package main

import "io"

type trackedReader struct {
	u    io.Reader
	size int64
	read int64
	ch   chan float32
}

func newTrackedReader(underlying io.Reader, size int64, progress chan float32) io.Reader {
	return &trackedReader{
		u:    underlying,
		size: size,
		ch:   progress,
	}
}

func (r *trackedReader) Read(p []byte) (n int, err error) {
	n, err = r.u.Read(p)
	if err != nil {
		return n, err
	}
	r.read += int64(n)

	go func() {
		r.ch <- float32((float64(r.read) / float64(r.size)) * 100.0)
	}()

	return n, nil
}
