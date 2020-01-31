package testing

import (
	"goparty/mp3"
	"testing"
)

func TestPlayMp3File(t *testing.T) {
	err := mp3.PlayMp3File("../songs/mdma.mp3")

	if err != nil {
		t.Error(err.Error())
	}
}
