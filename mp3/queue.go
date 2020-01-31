package mp3

import "github.com/faiface/beep"

//The code for this comes from the beep tutorial: https://github.com/faiface/beep/wiki/Making-own-streamers

//MusicQueue is a datastruct to add more songs to the streamer
type MusicQueue struct {
	streamers []beep.Streamer
}

//Add adds a new entry to the musicqueue
func (q *MusicQueue) Add(streamers ...beep.Streamer) {
	q.streamers = append(q.streamers, streamers...)
}

//Stream implements the streamer interface
func (q *MusicQueue) Stream(samples [][2]float64) (n int, ok bool) {
	// successfully filled already. We loop until all samples are filled.
	filled := 0
	for filled < len(samples) {
		// There are no streamers in the queue, so we stream silence.
		if len(q.streamers) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		// We stream from the first streamer in the queue.
		n, ok := q.streamers[0].Stream(samples[filled:])
		// If it's drained, we pop it from the queue, thus continuing with
		// the next streamer.
		if !ok {
			q.streamers = q.streamers[1:]
		}
		// We update the number of filled samples.
		filled += n
	}
	return len(samples), true
}

//Err trivial error implementation
func (q *MusicQueue) Err() error {
	return nil
}
