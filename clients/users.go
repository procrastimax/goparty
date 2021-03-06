package clients

import (
	"fmt"
	"sync"
)

//users is the map datastructure which contains all users (identified by their ip) and how many songs they currently added for downloading
var (
	users = make(map[string]*Properties)
	mutex = &sync.Mutex{}
)

//Properties keeps track of user added songs for downloading and playing
type Properties struct {
	DownloadingSongs int
	PlaylistSongs    int
	UserName         string
}

//AddSongDownload increment the counter of added songs of a user by 1
func AddSongDownload(ip string) {
	mutex.Lock()
	if _, ok := users[ip]; ok == true {
		users[ip].DownloadingSongs++
	} else {
		userName := GetUserNameToIP(ip)
		users[ip] = &Properties{UserName: userName}
		users[ip].DownloadingSongs = 1
	}
	mutex.Unlock()
}

//AddSongPlaylist increments the counter for songs the user has in the playlist
func AddSongPlaylist(ip string) {
	mutex.Lock()
	if _, ok := users[ip]; ok == true {
		users[ip].PlaylistSongs++
	} else {
		userName := GetUserNameToIP(ip)
		users[ip] = &Properties{UserName: userName}
		users[ip].PlaylistSongs = 1
	}
	mutex.Unlock()
}

//SongDoneDownloading decreases the song count of a user when the song is done downloading
func SongDoneDownloading(ip string) {
	mutex.Lock()
	if count, ok := users[ip]; ok == true {
		if count.DownloadingSongs > 0 {
			users[ip].DownloadingSongs--
		}
	}
	mutex.Unlock()
}

//SongDonePlaying decreases the counter of songs a user has in the playlist
func SongDonePlaying(ip string) {
	mutex.Lock()
	if count, ok := users[ip]; ok == true {
		if count.PlaylistSongs > 0 {
			users[ip].PlaylistSongs--
		}
	}
	mutex.Unlock()
}

//GetUserAddedSongs returns the number of added songs of a user, returns -1 if the user does not exist
func GetUserAddedSongs(ip string) *Properties {
	mutex.Lock()
	defer mutex.Unlock()
	if i, ok := users[ip]; ok == true {
		return i
	}
	return nil
}

//GetUserName returns the user name to the given IP
func GetUserName(ip string) string {
	if i, ok := users[ip]; ok == true {
		return i.UserName
	}
	return ""
}

//Count returns the size of the user map, can be used to see how many users added a song for downloading
func Count() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(users)
}

//GetUserCounts returns the user added map
func GetUserCounts() {
	for k := range users {
		fmt.Println(k, " - ", users[k].DownloadingSongs, " - ", users[k].PlaylistSongs)
	}
}
