package main

import (
	"fmt"
	"goparty/mp3"
	"goparty/youtube"
	"log"
)

const (
	songDir = "songs/"
)

func main() {
	testMusic()
	//testYoutube()
}

func testYoutube() {
	log.Println(youtube.DownloadYoutubeVideoAsMP3("https://www.youtube.com/watch?v=BWdPCGIzzuk", songDir, true))
}

func testMusic() {
	err := mp3.InitSpeaker()
	if err != nil {
		log.Fatalln(err.Error())
	}

	mp3.StartSpeaker()

	for {
		var name string
		fmt.Print("Type an MP3 file name: ")
		fmt.Scanln(&name)

		if name == "skip" {
			mp3.SkipSong()
			continue
		}

		if name == "quit" {
			mp3.CloseSpeaker()
			break
		}

		err := mp3.AddMP3ToMusicQueue(name)
		if err != nil {
			fmt.Println(err)
			break
		}

	}

	mp3.CloseSpeaker()
}
