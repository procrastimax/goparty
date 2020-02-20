package tests

import (
	"fmt"
	"goparty/mp3"
	"testing"
)

func TestInitializeSongDBFromMemory(t *testing.T) {

	err := mp3.InitializeSongDBFromMemory("../songs/")

	if err != nil {
		t.Error(err)
	}

	fmt.Println(mp3.GetSongDB())
}

func TestAddSongToDB(t *testing.T) {
	err := mp3.InitializeSongDBFromMemory(".")

	if err != nil {
		t.Error(err)
	}

	mp3.AddSongToDB("helloworld", "helloworld.mp3")

	if len(mp3.GetSongDB()) != 1 {
		t.Error("Song could not be added to emty songdb")
	}
	fmt.Println(mp3.GetSongDB())
}
