package mp3

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

/**
	Handling all downloaded songs, and songs inside a directory
**/

var (
	//SongDB represents a database of all offline available songs, whether through downloaded with youtube-dl or just an offline library
	songDB        map[string][]string
	mutex         sync.Mutex
	ytDownloadDir string
	songdir       string
)

//InitializeSongDBFromMemory initilaizes the songDB with all mp3 files from a given directory
//also traverses subdirectories
//this should happen at application start
func InitializeSongDBFromMemory(songDIr, downloadDir string) error {

	ytDownloadDir = downloadDir
	songdir = songDIr

	songDB = make(map[string][]string)
	songs := make([]string, 0)
	//dirPath keeps track of the folder which is currently iterated, when it change, we start a new list (bc of new folder)
	dirPath := ""
	err := filepath.Walk(songDIr,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("ReadSongsFromMemory: %s", err)
			}

			//only process files
			if info.IsDir() == false {
				tempPath := strings.TrimSuffix(path, info.Name())

				//dirPath is empty, set it
				if len(dirPath) == 0 {
					dirPath = tempPath
				} else if dirPath != tempPath {
					//if the dirPath differs from the current path, then we encountered a new directory
					//we need to save all current traversed songs to the dirPath and set dirPath to the new path
					songDB[dirPath] = songs
					songs = make([]string, 0)
					dirPath = tempPath
				}

				//only save .mp3 files
				if strings.HasSuffix(info.Name(), ".mp3") == true {
					songs = append(songs, info.Name())
				}
			}

			return nil
		})
	if err != nil {
		return fmt.Errorf("ReadSongsFromMemory: %s", err)
	}

	//if ended, then we need to add the last section
	songDB[dirPath] = songs

	return nil
}

//AddSongToDB adds a song to the database with the given songname and songpath
func AddSongToDB(songDir, songname string) {
	if songDB == nil {
		panic("SongDB Map not initialized! Call InitializeSongDBFromMemory!")
	}

	mutex.Lock()
	v, ok := songDB[songDir]
	if ok == true {
		songDB[songDir] = append(v, songname)
	} else {
		songDB[songDir] = []string{songname}
	}

	mutex.Unlock()
}

//CheckSongInDB returns a boolean indicating if the songname exists in the DB
func CheckSongInDB(songname string) bool {
	for _, v := range songDB {
		for _, song := range v {
			if strings.Contains(song, songname) {
				return true
			}
		}
	}
	return false
}

//CheckYTSongInDB checks whether or not a given song downloaded by youtubedl is in the map
//it returns the matching filename if found
func CheckYTSongInDB(ytURL string, downloadDir string) (string, error) {
	var videoIDStr string

	//we got a youtube shortform url
	isMatch, err := regexp.MatchString("https{0,1}://youtu\\.be/\\S*", ytURL)

	if err != nil {
		log.Fatalln("Regex for checking url against http://youtu.be/ link is invalid!")
	}

	if isMatch {
		strArr := strings.Split(ytURL, "/")
		if len(strArr) != 4 {
			return "", fmt.Errorf("provided youtube link has not supported format https://youtu.be/ID - %s", ytURL)
		}

		if strings.ContainsAny(strArr[3], "?") {
			videoIDStr = strings.Split(strArr[3], "?")[0]
		} else {
			videoIDStr = strArr[3]
		}

	} else {
		strArr := strings.Split(ytURL, "v=")
		if len(strArr) != 2 {
			return "", fmt.Errorf("provided youtube link has not supported format (?v=ID) - %s", ytURL)
		}

		if strings.ContainsAny(strArr[1], "&") {
			videoIDStr = strings.Split(strArr[1], "&")[0]
		} else {
			videoIDStr = strArr[1]
		}
	}

	songs, ok := songDB[downloadDir]
	if ok == false {

	}

	for _, song := range songs {
		if strings.Contains(song, videoIDStr) {
			log.Println("found: ", song)
			return song, nil
		}
	}

	//could not find the youtube song in the songdb
	return "", nil
}

//GetSongDB returns the songDB map
func GetSongDB() map[string][]string {
	return songDB
}

//GetSongDirAndCompleteName returns the directory of a given song, returns empty string when song is not in songDB
// also returns the real name of the song
func GetSongDirAndCompleteName(songname string) (string, string) {
	for k, v := range songDB {
		for _, song := range v {
			if strings.Contains(song, songname) {
				return k, song
			}
		}
	}
	return "", ""
}

//GetSortedSongList returns a list of all songnames sorted by their pathname, pathnames are also provided, recognizable by their '/' at the end
//this function also prettifies all songnames f.e. remove .mp3 suffix
func GetSortedSongList() []string {
	songDirKeys := make([]string, 0)

	songs := make([]string, 0)

	//we always want to have the yt directory at the top
	songDirKeys = append(songDirKeys, ytDownloadDir)

	for k := range songDB {
		if k != ytDownloadDir {
			songDirKeys = append(songDirKeys, k)
		}
	}

	//sort the keys
	sort.Strings(songDirKeys[1:])

	for _, elem := range songDirKeys {
		songs = append(songs, strings.TrimPrefix(elem, songdir))

		for _, songName := range songDB[elem] {
			//check if the songs has the #____# mark
			songName = strings.TrimSuffix(songName, ".mp3")

			if strings.Contains(songName, "#____#") {
				songName = strings.Split(songName, "#____#")[0]
			}
			//remove brackets
			songName = ParenthesisRegex.ReplaceAllString(songName, "")
			songs = append(songs, songName)
		}
	}
	return songs
}
