package youtube

import "sync"

//users is the map datastructure which contains all users (identified by their ip) and how many songs they currently added for downloading
var (
	users = make(map[string]int)
	mutex = &sync.Mutex{}
)

//UserAddSong increment the counter of added songs of a user by 1
func UserAddSong(ip string) {
	mutex.Lock()
	if _, ok := users[ip]; ok == true {
		users[ip]++
	} else {
		users[ip] = 1
	}
	mutex.Unlock()
}

//UserSongDone decreases the song count of a user when the song is done downloading
func UserSongDone(ip string) {
	mutex.Lock()
	if count, ok := users[ip]; ok == true {
		if count > 0 {
			users[ip]--
		}
	}
	mutex.Unlock()
}

//ResetUser sets the count of added songs by a user to 0
func ResetUser(ip string) {
	mutex.Lock()
	if _, ok := users[ip]; ok == true {
		users[ip] = 0
	}
	mutex.Unlock()
}

//ResetAllUser resets the count of added songs from all users to 0
func ResetAllUser() {
	mutex.Lock()
	users = make(map[string]int)
	mutex.Unlock()
}

//GetUserAddedSongs returns the number of added songs of a user, returns -1 if the user does not exist
func GetUserAddedSongs(ip string) int {
	mutex.Lock()
	defer mutex.Unlock()
	if i, ok := users[ip]; ok == true {
		return i
	}
	return 0
}

//UserCount returns the size of the user map, can be used to see how many users added a song for downloading
func UserCount() int {
	mutex.Lock()
	defer mutex.Unlock()
	return len(users)
}
