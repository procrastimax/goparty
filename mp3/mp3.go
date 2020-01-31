//Package mp3 handles the playing of mp3 files
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
}

//StartSpeaker starts the speaking by using the musicqueue
func StartSpeaker() {
	speaker.Play(&queue)
	fmt.Println("Speaker started")
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
		return fmt.Errorf("add mp3 queue: %v", err)
	}

	// we need to resample the song sample rate to the speaker sample rate
	resampledStreamer := beep.Resample(3, format.SampleRate, SampleRate, *streamer)
	speaker.Lock()
	queue.Add(resampledStreamer)
	speaker.Unlock()

	fmt.Printf("Added song to queue: %s\n", filename)
	return nil
}

//loadMp3File loads an mp3 file from the storage and returns it as a streamer and format
func loadMp3File(filename string) (*beep.StreamSeekCloser, *beep.Format, error) {

	if len(filename) <= 4 {
		return nil, nil, fmt.Errorf("load MP3: File %s is not a valid mp3 name", filename)
	}

	//check if really an mp3
	if filename[len(filename)-4:] != ".mp3" {
		return nil, nil, fmt.Errorf("load MP3: File %s is not a mp3", filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("load mp3: error when loading %s - %v", filename, err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return nil, nil, fmt.Errorf("play mp3: could not decode mp3 %v", err)
	}

	return &streamer, &format, nil
}
