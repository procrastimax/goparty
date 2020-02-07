package server

//users is the map datastructure which contains all users (identified by their ip) and how many songs they currently added
var users map[string]int

//UserAddSong increment the counter of added songs of a user by 1
func UserAddSong(ip string) {
	if _, ok := users[ip]; ok == true {
		users[ip]++
	} else {
		users[ip] = 1
	}
}

//ResetUser sets the count of added songs by a user to 0
func ResetUser(ip string) {
	if _, ok := users[ip]; ok == true {
		users[ip] = 0
	}
}

//ResetAllUser resets the count of added songs from all users to 0
func ResetAllUser() {
	users = make(map[string]int)
}

//GetUserAddedSongs returns the number of added songs of a user, returns -1 if the user does not exist
func GetUserAddedSongs(ip string) int {
	if i, ok := users[ip]; ok == true {
		return i
	}
	return -1
}

//UserCount returns the size of the user map, can be used to see how many users added a song
func UserCount() int {
	return len(users)
}
