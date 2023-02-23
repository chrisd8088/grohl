package grohl

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Statter describes the interface used by the g2s Statter object.
// http://godoc.org/github.com/peterbourgon/g2s
type Statter interface {
	Counter(sampleRate float32, bucket string, n ...int)
	Timing(sampleRate float32, bucket string, d ...time.Duration)
	Gauge(sampleRate float32, bucket string, value ...string)
}

// Counter writes a counter value to the Context.
func (c *Context) Counter(sampleRate float32, bucket string, n ...int) {
	//#nosec G404 -- Pseudo-random values are sufficient for sampling.
	if rand.Float32() > sampleRate {
		return
	}

	for _, num := range n {
		c.Log(Data{"metric": bucket, "count": num})
	}
}

// Timing writes a timer value to the Context.
func (c *Context) Timing(sampleRate float32, bucket string, d ...time.Duration) {
	//#nosec G404 -- Pseudo-random values are sufficient for sampling.
	if rand.Float32() > sampleRate {
		return
	}

	for _, dur := range d {
		c.Log(Data{"metric": bucket, "timing": int64(dur / time.Millisecond)})
	}
}

// Gauge writes a static value to the Context.
func (c *Context) Gauge(sampleRate float32, bucket string, value ...string) {
	//#nosec G404 -- Pseudo-random values are sufficient for sampling.
	if rand.Float32() > sampleRate {
		return
	}

	for _, v := range value {
		c.Log(Data{"metric": bucket, "gauge": v})
	}
}

// Embedded in Context and Timer.
type _statter struct {
	statter           Statter
	StatterSampleRate float32
	StatterBucket     string
}

// SetStatter sets a Statter to be used in Timer Log() calls.
func (s *_statter) SetStatter(statter Statter, sampleRate float32, bucket string) {
	s.statter = statter
	s.StatterSampleRate = sampleRate
	s.StatterBucket = bucket
}

// StatterBucketSuffix changes the suffix of the bucket.  If SetStatter() is
// called with bucket of "foo", then StatterBucketSuffix("bar") changes it to
// "foo.bar".
func (s *_statter) StatterBucketSuffix(suffix string) {
	if len(s.StatterBucket) == 0 {
		s.StatterBucket = suffix
		return
	}

	sep := "."
	if strings.HasSuffix(s.StatterBucket, ".") {
		sep = ""
	}
	s.StatterBucket += fmt.Sprintf("%s%s", sep, suffix)
}

// Timing sends the timing to the configured Statter.
func (s *_statter) Timing(dur time.Duration) {
	if s.statter == nil {
		s.statter = CurrentStatter
	}

	s.statter.Timing(s.StatterSampleRate, s.StatterBucket, dur)
}

func (s *_statter) dup() *_statter {
	return &_statter{s.statter, s.StatterSampleRate, s.StatterBucket}
}
