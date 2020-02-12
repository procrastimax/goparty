package user

import (
	"io/ioutil"
	"strconv"
	"strings"
)

var (
	usernames []string
)

//GetAllUserNames returns a list of all available user names
func GetAllUserNames() []string {
	return usernames
}

//GetUserNameToIP returns a string matching a username for a specific number
func GetUserNameToIP(ip string) string {
	split := strings.Split(ip, ":")
	if len(split) > 1 {
		ip = split[0]
	}
	//get last number of IP which should be in a range of 0-254
	lastNum, err := strconv.Atoi(strings.Split(ip, ".")[3])
	if err != nil {
		return ip
	}
	if lastNum < len(usernames) {
		return usernames[lastNum]
	}
	return ip
}

//InitUserNames reads in the username file and stores them internally
func InitUserNames(filename string) error {
	//read in username file
	fileContent, err := ioutil.ReadFile("usernames.txt")
	if err != nil {
		return err
	}
	usernames = strings.Split(string(fileContent), "\n")
	return nil
}
