//Package youtube handles the downloading of youtube videos and their conversion to mp3
package youtube

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

//DownloadYoutubeVideoAsMP3 downloads a youtube video in mp3 format
func DownloadYoutubeVideoAsMP3(url string, songDir string, verbose bool) error {
	youtubeDL, err := exec.LookPath("youtube-dl")
	if err != nil {
		return fmt.Errorf("download yt mp3: youtube-dl is not installed/ findable in $PATH")
	}
	//weird that the output format get strangely parsed... "-osongs/"" should be "-o songs/""
	cmd := exec.Command(youtubeDL, "-i", "--flat-playlist", "--no-playlist", "--extract-audio", "--youtube-skip-dash-manifest", "--audio-format=mp3", "-o"+songDir+"/%(title)s.%(ext)s", url)

	var stderr bytes.Buffer

	if verbose {
		cmd.Stdout = os.Stdout
	}

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errStr := string(stderr.Bytes())
		return fmt.Errorf("download yt mp3: %s - %v", errStr, err)
	}
	return nil
}
