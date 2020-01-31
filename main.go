package main

import (
	"fmt"
	"goparty/mp3"
	"log"
)

func main() {

	err := mp3.InitSpeaker()
	if err != nil {
		log.Fatalln(err.Error())
	}

	mp3.StartSpeaker()

	for {
		var name string
		fmt.Print("Type an MP3 file name: ")
		fmt.Scanln(&name)

		err := mp3.AddMP3ToMusicQueue(name)
		if err != nil {
			fmt.Println(err)
			break
		}

	}

	mp3.CloseSpeaker()
}
