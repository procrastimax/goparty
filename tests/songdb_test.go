package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/procrastimax/goparty/mp3"
)

func TestInitializeSongDBFromMemory(t *testing.T) {

	err := mp3.InitializeSongDBFromMemory("../songs/", "../songs/yt")

	if err != nil {
		t.Error(err)
	}

	songList := mp3.GetSortedSongList()

	for _, entry := range songList {
		if strings.HasSuffix(entry, "/") {
			fmt.Println(entry)
		} else {
			fmt.Println("\t", entry)
		}
	}
}

func TestAddSongToDB(t *testing.T) {
	err := mp3.InitializeSongDBFromMemory(".", ".")

	if err != nil {
		t.Error(err)
	}

	mp3.AddSongToDB("helloworld/", "helloworld")

	if len(mp3.GetSongDB()) != 1 {
		t.Error("Song could not be added to emty songdb")
	}
	fmt.Println(mp3.GetSongDB())
}

func TestGetSongDir(t *testing.T) {
	err := mp3.InitializeSongDBFromMemory(".", ".")

	if err != nil {
		t.Error(err)
	}

	mp3.AddSongToDB("helloworld/", "helloworld.mp3")

	if len(mp3.GetSongDB()) != 1 {
		t.Error("Song could not be added to emty songdb")
	}

	songDir, songname := mp3.GetSongDirAndCompleteName("helloworld")
	if len(songDir) == 0 || songname != "helloworld.mp3" {
		t.Error("Could not get directory of song!")
	}

	fmt.Println("dir -->", songDir)
}
