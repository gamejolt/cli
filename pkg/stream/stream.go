package stream

import (
	"io"
	"time"
)

// Sample is a sample reporting the rate of data passing through the stream
type Sample struct {
	SampleID      int64 `json:"sampleId"`
	Current       int64 `json:"bytesPerSec"`
	Peak          int64 `json:"peak"`
	Low           int64 `json:"low"`
	Average       int64 `json:"average"`
	MovingPeak    int64 `json:"movingPeak"`
	MovingLow     int64 `json:"movingLow"`
	MovingAverage int64 `json:"movingAverage"`
}

// SpeedSampler is a reader that reports speed samples for the stream its reading from
type SpeedSampler struct {
	r                io.Reader
	callback         SampleCallback
	samplesPerSecond int64
	samplesTaken     int64
	ticker           *time.Ticker
	samples          *ring
	total            int64
	current          int64
	peak             int64
	low              int64
	average          int64
	movingPeak       int64
	movingLow        int64
	movingAverage    int64
}

// SampleCallback is the callback the SpeedSampler calls for every sample
type SampleCallback func(sampler *SpeedSampler, sample *Sample)

// NewSpeedSampler creates a new SpeedSampler
func NewSpeedSampler(r io.Reader, callback SampleCallback) (*SpeedSampler, error) {
	sampler := &SpeedSampler{
		r:                r,
		callback:         callback,
		samplesPerSecond: 2,
	}

	if err := sampler.Start(); err != nil {
		return nil, err
	}

	return sampler, nil
}

// Reset resets the SpeedSampler. The stream will still be readable, but no samples will be reported.
func (s *SpeedSampler) Reset() {
	if s.ticker != nil {
		s.ticker.Stop()
		s.ticker = nil
	}

	if s.samples != nil {
		s.samples = nil
	}
	s.samplesTaken = 0

	s.current, s.peak, s.average, s.movingPeak, s.movingAverage = 0, 0, 0, 0, 0
	s.low, s.movingLow = -1, -1
}

// Start starts the SpeedSampler reporting.
func (s *SpeedSampler) Start() error {
	s.Reset()

	s.samples = newRing(s.samplesPerSecond * 5)
	s.ticker = time.NewTicker(time.Duration(1000/s.samplesPerSecond) * time.Millisecond)
	go func() {
		for {
			if s.ticker == nil {
				break
			}

			_, open := <-s.ticker.C
			if !open {
				break
			}

			s.takeSample()
		}
	}()

	return nil
}

func (s *SpeedSampler) takeSample() *Sample {
	s.samplesTaken++

	s.current *= s.samplesPerSecond
	s.average += (s.current - s.average) / s.samplesTaken
	s.peak = max(s.peak, s.current)
	if s.low == -1 {
		s.low = s.current
	} else {
		s.low = min(s.low, s.current)
	}

	sample := &Sample{
		SampleID: s.samplesTaken,
		Current:  s.current,
		Peak:     s.peak,
		Low:      s.low,
		Average:  s.average,
	}

	s.samples.add(sample)
	s.movingPeak, s.movingLow, s.movingAverage = 0, -1, 0
	for i := int64(0); i < s.samples.cl; i++ {
		tempSample := s.samples.get(i).(*Sample)
		s.movingPeak = max(s.movingPeak, tempSample.Current)
		if s.movingLow == -1 {
			s.movingLow = tempSample.Current
		} else {
			s.movingLow = min(s.movingLow, tempSample.Current)
		}
		s.movingAverage += tempSample.Current
	}
	s.movingAverage /= min(s.samples.cl, s.samples.l)
	sample.MovingPeak = s.movingPeak
	sample.MovingLow = s.movingLow
	sample.MovingAverage = s.movingAverage

	s.current = 0
	if s.callback != nil {
		s.callback(s, sample)
	}
	return sample
}

// Sample returns the last speed sample the SpeedSampler created
func (s *SpeedSampler) Sample() *Sample {
	if s.samples == nil || s.samples.cl == 0 {
		return nil
	}
	return s.samples.get(0).(*Sample)
}

// Total returns the total amount of bytes read so far.
func (s *SpeedSampler) Total() int64 {
	return s.total
}

func (s *SpeedSampler) Read(p []byte) (n int, err error) {
	n, err = s.r.Read(p)
	s.current += int64(n)
	s.total += int64(n)
	return n, err
}

// Close closes the underlying reader. This implements ReadCloser
func (s *SpeedSampler) Close() error {
	s.Reset()
	if r, ok := s.r.(io.ReadCloser); ok {
		r.Close()
	}
	return nil
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b int64) int64 {
	if a <= b {
		return a
	}
	return b
}

type ring struct {
	e  []interface{}
	l  int64
	cl int64
	c  int64
	n  int64
}

// Simple cyclic ring buffer to keep the samples in
func newRing(length int64) *ring {
	return &ring{
		e:  make([]interface{}, length, length),
		l:  length,
		cl: 0,
		c:  0,
		n:  0,
	}
}

func (r *ring) add(values ...interface{}) {
	for _, value := range values {
		r.e[r.n] = value
		if r.cl == r.l {
			r.c = (r.c + 1) % r.l
		}
		r.cl = min(r.cl+1, r.l)
		r.n = (r.n + 1) % r.l
	}
}

func (r *ring) get(i int64) interface{} {
	return r.e[(r.c+i)%r.l]
}
