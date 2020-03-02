package mp3

import (
	"fmt"
	"strings"
	"sync"

	"github.com/faiface/beep"
	"github.com/procrastimax/goparty/clients"
)

var (
	neededUpvoteCount = 1
)

//SetNeededUpvoteCount sets the number of upvotes needed for a song to consider a reranking of the queue
func SetNeededUpvoteCount(upvoteCount int) {
	neededUpvoteCount = upvoteCount
}

//The code for this queue comes from the beep tutorial: https://github.com/faiface/beep/wiki/Making-own-streamers

//Song representing a single song from the downloaded queue of songs
type Song struct {
	SongName  string
	UserIP    string
	UserName  string
	SongCount int
	//upvotes is a list of strings, each string represents a userIP which upvoted the song
	upvotes []string
}

//Upvote adds a user who upvoted the song to the upvotes list
func (s *Song) Upvote(userIP string) {
	//check if the upvotes list got initialized before
	//because most likely we dont need the list
	if s.upvotes == nil {
		// upvotes list is nil, we need to initialize it
		s.upvotes = make([]string, 0)
	}
	for _, elem := range s.upvotes {
		if elem == userIP {
			//user already upvoted the song, do nothing
			return
		}
	}
	s.upvotes = append(s.upvotes, userIP)
}

//GetUpvotes returns a list of all users who upvoted the song
func (s *Song) GetUpvotes() []string {
	return s.upvotes
}

//GetUpvotesCount returns the number of upvotes for the song
func (s *Song) GetUpvotesCount() int {
	return len(s.upvotes)
}

func (s Song) String() string {
	return fmt.Sprintf("%s - %s -> %s : %d", s.SongName, *getOnlyIP(&s.UserIP), clients.GetUserName(s.UserIP), s.SongCount)
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
	sync.Mutex
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
	q.Lock()
	clients.AddSongPlaylist(userIP)

	songStream := songStream{
		Song{SongName: songame,
			SongCount: clients.GetUserAddedSongs(userIP).PlaylistSongs,
			UserIP:    userIP,
			UserName:  clients.GetUserName(userIP)},
		&streamer,
	}

	//like in the downloading section, add the song at the position where the count of added songs differ from the next one
	if len(q.songs) <= 1 {
		q.songs = append(q.songs, songStream)
	} else {
		startValue := clients.GetUserAddedSongs(userIP).PlaylistSongs
		for i, val := range q.songs {
			if val.SongCount > startValue {
				//when the following song has more upvotes, then skip it
				//and check next song
				if val.GetUpvotesCount() > 0 {
					break
				}
				//Insert element at position 'i'
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
	q.Unlock()
}

//UpvoteSong adds a user to the upvoted song specified by the songID which is the current ID of the song in the queue
func (q *MusicQueue) UpvoteSong(songID int, userIP string) {
	q.Lock()
	defer q.Unlock()
	q.songs[songID].Upvote(userIP)

	//after upvoting the song we want to decrease the added count, so the song moves forward in the queue
	//therefore we need to check the element before the upvoted song

	//check if song is already first element or 2nd, if this is the case then do nothing. Also when the upvote count of the song is less than the needed upvote count, then also just return
	if songID <= 1 || q.songs[songID].GetUpvotesCount() < neededUpvoteCount {
		return
	}

	//check song before current song, if the song before this song has a different SongCount value
	//then we need to decrease the songcount for this song so when adding new songs we have coherent values
	if q.songs[songID-1].SongCount == q.songs[songID].SongCount {
		q.songs[songID].SongCount--
	}

	//only swap songs, when the previous song has less upvotes than the current one
	if q.songs[songID-1].GetUpvotesCount() >= q.songs[songID].GetUpvotesCount() {
		return
	}

	temp := q.songs[songID-1]
	q.songs[songID-1] = q.songs[songID]
	q.songs[songID] = temp
}

//GetUpvotesForSong returns the upvotes for a given songID
func (q *MusicQueue) GetUpvotesForSong(songID int) []string {
	return q.songs[songID].GetUpvotes()
}

//CheckUserUpvotedSong returns true if the given userIP already upvoted the song with the given songID
func (q *MusicQueue) CheckUserUpvotedSong(songID int, userIP string) bool {
	for _, elem := range q.songs[songID].upvotes {
		if elem == userIP {
			return true
		}
	}
	return false
}

//Done skips to the next song
func (q *MusicQueue) Done() {
	q.Lock()
	if len(q.songs) > 0 {
		userIP := q.songs[0].UserIP
		clients.SongDonePlaying(userIP)
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
	q.Unlock()
}

//Pause pauses the music
func (q *MusicQueue) Pause() {
	q.isPaused = true
}

//Resume resumes music
func (q *MusicQueue) Resume() {
	q.isPaused = false
}

//Clear deletes all entries in the music queue
func (q *MusicQueue) Clear() {
	q = &MusicQueue{}
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
