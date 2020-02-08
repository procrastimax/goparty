//Package youtube handles the downloading of youtube videos and their conversion to mp3
package youtube

/*
	Currently all youtube songs which need to be downloaded get queued in in our DownloadQueue
	When the queue is empty, it sends a url over the channel to activate a downloading toolchain.
	When the song is downloaded and converted, thena callback function of the queue is called.
	In this callback function we check if the queue is empty or not. If the queue is not empty,
	then we pass another url to the channel to start the download again. If the queue is empty,
	nothing happens.

	Maybe improve this in the future by using more then one download job concurrently. But the basic
	idea behind the current implementation is, that while playing the first downloaded song, all other
	are able to download more songs and the machine has time to download those other songs.
*/

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

var (
	downloadDir  = "songs/"
	youtubeDlDir = ""
	isVerbose    = false
	jobCh        = make(chan downloadEntity, 2)
	quitCh       = make(chan bool)
	queue        downloadQueue
)

type downloadEntity struct {
	url        string
	userIP     string
	addedCount int
}

func (d downloadEntity) String() string {
	return fmt.Sprintf("%s - %s -> %d", d.url, d.userIP, d.addedCount)
}

//downloadQueue handles information about upcoming songs to download
type downloadQueue struct {
	songs []downloadEntity
	sync.Mutex
}

//Add adds an url to the worker list of urls
func Add(url string, userIP string) {
	queue.Lock()

	UserAddSong(userIP)

	song := downloadEntity{url, userIP, GetUserAddedSongs(userIP)}

	if len(queue.songs) <= 1 {
		queue.songs = append(queue.songs, song)
	} else {
		//insert the song in the queue, at this position, where the addedCount increases
		//all songs in the beginning of the queue have a addedCount of 1
		startValue := 1
		for i, val := range queue.songs {
			if val.addedCount != startValue {
				//create copy of last element and append it to queue
				queue.songs = append(queue.songs, queue.songs[len(queue.songs)-1])
				copy(queue.songs[i+1:], queue.songs[i:len(queue.songs)-1])
				queue.songs[i] = song
				break
			}
			//when we haven't found a change yet, then also just append the song
			if i == len(queue.songs)-1 {
				queue.songs = append(queue.songs, song)
			}
		}
	}

	if len(queue.songs) == 1 {
		jobCh <- song
	}
	queue.Unlock()
}

//ExitDownloadWorker quits the donloading worker loop by sending a value on the quit channel
func ExitDownloadWorker() {
	quitCh <- true
}

//done removes the first element of the queue when done, also decreases the addedcount of the user by 1 for all added songs
func done(userIP string) {
	queue.Lock()
	UserSongDone(userIP)

	for i, val := range queue.songs {
		if val.userIP == userIP {
			if val.addedCount > 0 {
				queue.songs[i].addedCount--
			}
		}
	}

	queue.songs = queue.songs[1:]
	if len(queue.songs) != 0 {
		jobCh <- queue.songs[0]
	}
	queue.Unlock()
}

//MustExistYoutubeDL is a helper function, which panics when no youtube-dl exist
func MustExistYoutubeDL() {
	dir, err := exec.LookPath("youtube-dl")
	if err != nil {
		panic("youtube-dl is not installed - cannot find it in $PATH")
	}
	youtubeDlDir = dir
}

//StartDownloadWorker starts downlading
func StartDownloadWorker() {
	fmt.Println("Started YT-Download Worker!")
	var err error
	go func() {
		for {
			select {
			case <-quitCh:
				fmt.Println("Stopping Download Worker")
				return

			case job := <-jobCh:
				fmt.Println("received job: ", job)
				err = downloadYoutubeVideoAsMP3(&job, downloadDir, isVerbose, done)
				if err != nil {
					log.Fatalln(err)
				}
			}
		}
	}()
}

//downloadYoutubeVideoAsMP3 downloads a youtube video in mp3 format
func downloadYoutubeVideoAsMP3(song *downloadEntity, downloadDir string, verbose bool, callback func(userIP string)) error {
	if len(youtubeDlDir) == 0 {
		panic("youtube-dl directory variable was not set previously!")
	}

	defer callback(song.userIP)

	//weird that the output format get strangely parsed... "-osongs/"" should be "-o songs/""
	cmd := exec.Command(youtubeDlDir, "-i", "--flat-playlist", "--no-playlist", "--extract-audio", "--youtube-skip-dash-manifest", "--audio-format=mp3", "-o"+downloadDir+"/%(title)s___%(id)s___.%(ext)s", song.url)
	var stderr bytes.Buffer

	if verbose {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("%s", errStr)
	}
	return nil
}
