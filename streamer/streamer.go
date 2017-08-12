// Package streamer provides the time-scale modification methods as Streamers, to be used with the beep library (https://github.com/faiface/beep)
package streamer

import "github.com/faiface/beep"

type Streamer struct {
	Streamer beep.Streamer
}

func (s Streamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = s.Streamer.Stream(samples)
	return n, ok
}

// Err propagates the wrapped Streamer's errors.
func (s Streamer) Err() error {
	return s.Streamer.Err()
}
