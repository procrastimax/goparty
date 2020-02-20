package mp3

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

/**
	Handling all downloaded songs, and songs inside a directory
**/

var (
	//SongDB represents a database of all offline available songs, whether through downloaded with youtube-dl or just an offline library
	//the key must be the songname, the value is the songpath
	songDB map[string]string
	mutex  sync.Mutex
)

//InitializeSongDBFromMemory initilaizes the songDB with all mp3 files from a given directory
//also traverses subdirectories
//this should happen at application start
func InitializeSongDBFromMemory(songDIr string) error {
	songDB = make(map[string]string)
	err := filepath.Walk(songDIr,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("ReadSongsFromMemory: %s", err)
			}

			//only process files
			if info.IsDir() == false {
				//only save .mp3 files
				if strings.HasSuffix(info.Name(), ".mp3") == true {
					songName := info.Name()
					//check if the songs have the #____# mark
					if strings.Contains(songName, "#____#") {
						songName = strings.Split(info.Name(), "#____#")[0]
						//remove brackets
						songName = ParenthesisRegex.ReplaceAllString(songName, "")
					}
					songDB[songName] = path
				}
			}

			return nil
		})

	if err != nil {
		return fmt.Errorf("ReadSongsFromMemory: %s", err)
	}

	return nil
}

//AddSongToDB adds a song to the database with the given songname and songpath
func AddSongToDB(songname, songpath string) {
	mutex.Lock()
	_, ok := songDB[songname]
	//song already exists in songdb, but also overrind its value, because maybe something changed
	if ok {
		//log.Println("song already exsists in songDB, though adding it")
	}
	songDB[songname] = songpath
	mutex.Unlock()
}

//CheckSongInDB returns a boolean indicating if the songname exists in the DB
func CheckSongInDB(songname string) bool {
	_, ok := songDB[songname]
	return ok
}

//GetSongPath returns the path of a given song and the information whether or not the song exists in the db
func GetSongPath(songname string) (string, bool) {
	path, ok := songDB[songname]
	if ok {
		return path, true
	}
	return "", false
}

//GetFileDirAndFileName returns the file directory and filename of the given song
func GetFileDirAndFileName(songname string) (string, string) {
	songpath, ok := songDB[songname]
	if ok == false {
		return "", ""
	}
	splitArr := strings.Split(songpath, "/")
	filename := splitArr[len(splitArr)-1]
	filedir := strings.Join(splitArr[:len(splitArr)-1], "/")

	return filedir + "/", filename
}

//GetSongDB returns the songDB map
func GetSongDB() map[string]string {
	return songDB
}

//GetSortedSongNameList returns a list of all songnames sorted by their name
func GetSortedSongNameList() []string {
	var songList []string
	for k := range songDB {
		songList = append(songList, k)
	}
	sort.Strings(songList)
	return songList
}
