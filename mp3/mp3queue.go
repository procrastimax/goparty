package mp3

import (
	"fmt"
	"goparty/user"
	"strings"

	"github.com/faiface/beep"
)

//The code for this queue comes from the beep tutorial: https://github.com/faiface/beep/wiki/Making-own-streamers

//Song representing a single song from the downloaded queue of songs
type Song struct {
	SongCount int
	SongName  string
	UserIP    string
}

func (s Song) String() string {
	return fmt.Sprintf("%s - %s : %d", s.SongName, *getOnlyIP(&s.UserIP), s.SongCount)
}

func getOnlyIP(ip *string) *string {
	split := strings.Split(*ip, ":")
	if len(split) > 1 {
		return &split[0]
	}
	return ip
}

//songStream is a basic song with extended stream field
type songStream struct {
	Song
	streamer *beep.Streamer
}

//MusicQueue is a datastruct to add more songs to the streamer
// we dont need a mutex here, because this queue in only on the server side and handles one sound ouput
type MusicQueue struct {
	songs    []songStream
	isPaused bool
	currIdx  int
}

//GetSongs returns all songs from the music queue without streamer
func (q *MusicQueue) GetSongs() []Song {
	songs := make([]Song, len(q.songs))
	for i := range songs {
		songs[i] = q.songs[i].Song
	}
	return songs
}

//Add adds a new entry to the musicqueue
func (q *MusicQueue) Add(songame string, userIP string, streamer beep.Streamer) {
	user.AddSongPlaylist(userIP)

	songStream := songStream{
		Song{SongName: songame, SongCount: user.GetUserAddedSongs(userIP).PlaylistSongs, UserIP: userIP},
		&streamer,
	}
	//like in the downloading section, add the song at the position where the count of added songs differ from the next one
	if len(q.songs) <= 1 {
		q.songs = append(q.songs, songStream)
	} else {
		startValue := user.GetUserAddedSongs(userIP).PlaylistSongs
		for i, val := range q.songs {
			if val.SongCount > startValue {
				//Insert element at position 'i'
				//TODO: maybe dont actually swap the streamers, but instead the index for streamers in a seperate list
				q.songs = append(q.songs, q.songs[len(q.songs)-1])
				copy(q.songs[i+1:], q.songs[i:len(q.songs)-1])
				q.songs[i] = songStream
				break
			}
			//when we haven't found a change yet, then also just append the song, f.e. when all songs have count of 1
			if i == len(q.songs)-1 {
				q.songs = append(q.songs, songStream)
			}
		}
	}
}

//Done skips to the next song
func (q *MusicQueue) Done() {
	if len(q.songs) > 0 {
		userIP := q.songs[0].UserIP
		user.SongDonePlaying(userIP)
		q.songs = q.songs[1:]
		q.currIdx++
		//we need to iterate over the complete queue, and decrease the count of the user added songs
		for i := range q.songs {
			if q.songs[i].UserIP == userIP {
				if q.songs[i].SongCount > 0 {
					q.songs[i].SongCount--
				}
			}
		}
	}
}

//Pause pauses the music
func (q *MusicQueue) Pause() {
	q.isPaused = true
}

//Resume resumes music
func (q *MusicQueue) Resume() {
	q.isPaused = false
}

//Stream implements the streamer interface
func (q *MusicQueue) Stream(samples [][2]float64) (n int, ok bool) {
	// successfully filled already. We loop until all samples are filled.
	filled := 0
	for filled < len(samples) {
		// There are no streamers in the queue, so we stream silence.
		// If the isPaused flag is set, we also stream silence
		if len(q.songs) == 0 || q.isPaused {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		// We stream from the first streamer in the queue.
		n, ok := (*q.songs[0].streamer).Stream(samples[filled:])
		// If it's drained, we pop it from the queue, thus continuing with
		// the next streamer.
		if !ok {
			q.Done()
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
