package main

import (
	"math"
	"time"
)

type backoff struct {
	duration    time.Duration
	maxDuration time.Duration
	attempts    int
}

func (b *backoff) sleep() {
	b.attempts++

	d := b.duration * time.Duration(math.Pow(2, float64(b.attempts-1)))
	if b.maxDuration > 0 && d > b.maxDuration {
		d = b.maxDuration
	}

	time.Sleep(d)
}
