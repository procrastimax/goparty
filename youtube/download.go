//Package youtube handles the downloading of youtube videos and their conversion to mp3
package youtube

/*
	Currently all youtube songs which need to be downloaded get queued in in our DownloadQueue
	When the queue is empty, it sends a url over the channel to activate a downloading toolchain.
	When the song is downloaded and converted, thena callback function of the queue is called.
	In this callback function we check if the queue is empty or not. If the queue is not empty,
	then we pass another url to the channel to start the download again. If the queue is empty,
	nothing happens.

	TODO: Maybe improve this in the future by using more then one download job concurrently. But the basic
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
	isVerbose    = true
	jobCH        = make(chan string)
)

//DownloadQueue handles information about upcoming songs to download
type DownloadQueue struct {
	urls []string
	sync.Mutex
}

//Add adds an url to the worker list of urls
func (q *DownloadQueue) Add(url string) {
	q.Lock()
	q.urls = append(q.urls, url)

	if len(q.urls) == 1 {
		jobCH <- url
		fmt.Println("Added first element")
	}
	fmt.Println("download queue length:", len(q.urls))
	q.Unlock()
}

//Done removes the first element of the queue when done
func (q *DownloadQueue) Done() {
	q.Lock()
	fmt.Println("done")
	q.urls = q.urls[1:]
	if len(q.urls) != 0 {
		nextURL := q.urls[0]
		jobCH <- nextURL
	}
	fmt.Println("download queue length:", len(q.urls))
	q.Unlock()
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
func (q *DownloadQueue) StartDownloadWorker() {
	//wg := sync.WaitGroup{}
	go func() {
		for job := range jobCH {
			fmt.Println("hanging here before")
			log.Fatalln(downloadYoutubeVideoAsMP3(job, downloadDir, isVerbose, q.Done))
		}
	}()
}

//downloadYoutubeVideoAsMP3 downloads a youtube video in mp3 format
func downloadYoutubeVideoAsMP3(url string, downloadDir string, verbose bool, callback func()) error {
	if len(youtubeDlDir) == 0 {
		panic("youtube-dl directory variable was not set previously!")
	}

	defer callback()

	//weird that the output format get strangely parsed... "-osongs/"" should be "-o songs/""
	cmd := exec.Command(youtubeDlDir, "-i", "--flat-playlist", "--no-playlist", "--extract-audio", "--youtube-skip-dash-manifest", "--audio-format=mp3", "-o"+downloadDir+"/%(title)s___%(id)s___.%(ext)s", url)

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
