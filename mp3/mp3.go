//Package mp3 handles the playing and concatenation of mp3 streams
package mp3

import (
	"fmt"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	//SampleRate is the default speaker sample rate, currently we use the standard CD sample rate
	SampleRate = 44100
)

var (
	queue MusicQueue
)

//DeleteMusicQueue deletes all currently active streamer on the speaker
func DeleteMusicQueue() {
	speaker.Clear()
}

//CloseSpeaker closes the speaker
func CloseSpeaker() {
	DeleteMusicQueue()
	speaker.Close()
	fmt.Println("Speaker closed")
}

//PauseSpeaker pauses the speaker
func PauseSpeaker() {
	fmt.Println("Speaker paused")
	queue.Pause()
}

//StartSpeaker starts the speaking by using the musicqueue
func StartSpeaker() {
	//if queue is already initialized, then just resume playing
	if len(queue.streamers) > 0 {
		fmt.Println("Speaker resumed")
		queue.Resume()
	} else {
		speaker.Play(&queue)
		fmt.Println("Speaker started")
	}
}

//InitSpeaker initializes the speaker with a fixed sample rate
func InitSpeaker() error {
	sr := beep.SampleRate(SampleRate)
	err := speaker.Init(sr, sr.N(time.Second/10))

	if err != nil {
		return fmt.Errorf("init speaker: %v", err)
	}
	fmt.Println("Speaker initialized")
	return nil
}

//AddMP3ToMusicQueue adds a mp3 stream to the running music queue
func AddMP3ToMusicQueue(filename string) error {
	streamer, format, err := loadMp3File(filename)

	if err != nil {
		return fmt.Errorf("load mp3: %v", err)
	}

	// we need to resample the song sample rate to the speaker sample rate
	resampledStreamer := beep.Resample(3, format.SampleRate, SampleRate, *streamer)
	speaker.Lock()
	queue.Add(filename, resampledStreamer)
	speaker.Unlock()

	fmt.Printf("Added song to queue: %s\n", filename)
	return nil
}

//SkipSong skips a song in the music queue
func SkipSong() {
	queue.Skip()
	fmt.Println("Song skipped")
}

//loadMp3File loads an mp3 file from the storage and returns it as a streamer and format
func loadMp3File(filename string) (*beep.StreamSeekCloser, *beep.Format, error) {
	if len(filename) <= 4 {
		return nil, nil, fmt.Errorf("File %s is not a valid mp3 name", filename)
	}

	//check if really an mp3
	if filename[len(filename)-4:] != ".mp3" {
		return nil, nil, fmt.Errorf("File %s is not a mp3", filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, nil, err
	}

	return &streamer, &format, nil
}
