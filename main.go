package main

import (
	"fmt"
	"log"
	"yt-queue/mp3"
	"yt-queue/server"
	"yt-queue/youtube"
)

const (
	songDir = "songs/"
)

func main() {
	youtube.MustExistYoutubeDL()
	testServer()
	//testMusic()
}

func testServer() {
	server.SetupServing()
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
		} else if name == "pause" {
			mp3.PauseSpeaker()
			continue
		} else if name == "resume" {
			mp3.ResumeSpeaker()
			continue
		} else if name == "quit" {
			mp3.CloseSpeaker()
			break
		} else {
			err := mp3.AddMP3ToMusicQueue(name)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}
	mp3.CloseSpeaker()
}
