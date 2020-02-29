package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

//Config handles multiple settings used for the logic/ handling of the music queue
type Config struct {
	//DownloadPath specifies the location of the downloaded youtube song
	DownloadPath string `json:"downloadPath"`

	//MusicPath specifies the location used for creating an offline song database
	MusicPath string `json:"musicPath"`

	//UpvotesNeededForRanking specifies the count of upvotes for a song needed to change its position in the current playing queue
	UpvotesNeededForRanking int `json:"upvotesForRerank"`

	//AllUserAdmin when set to true, then all user receive the admin user interface and can stop/skip music
	AllUserAdmin bool `json:"allUserAdmin"`
}

//CreateInitialConfig creates the initial config if it wasn't created before.
func CreateInitialConfig(configPath string) error {
	if checkConfigExists(configPath) == false {
		fileName := strings.Split(configPath, string(os.PathSeparator))
		err := os.MkdirAll(strings.TrimSuffix(configPath, fileName[len(fileName)-1]), 0700)

		config := Config{
			DownloadPath:            "",
			MusicPath:               "",
			AllUserAdmin:            false,
			UpvotesNeededForRanking: 2,
		}

		file, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			return fmt.Errorf("CreateInitialConfig: %s", err)
		}

		err = ioutil.WriteFile(configPath, file, 0644)
		if err != nil {
			return fmt.Errorf("CreateInitialConfig: %s", err)
		}
	}
	return nil
}

func checkConfigExists(configPath string) bool {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return false
	}
	return true
}

//ReadConfig reads the current config and returns the config struct and a potential error
func ReadConfig(configPath string) (*Config, error) {
	if checkConfigExists(configPath) == false {
		//when file does not exist, create it first
		CreateInitialConfig(configPath)
	}

	file, err := ioutil.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("ReadConfig: %s", err)
	}

	config := &Config{}

	err = json.Unmarshal(file, config)

	if err != nil {
		return nil, fmt.Errorf("ReadConfig: %s", err)
	}

	if len(config.DownloadPath) == 0 {
		log.Printf("\n\n-->It seems that you haven't already set a path for specifing in which folder the youtube songs should be downloaded!\nPlease set one under %s\n\n", configPath)
		return nil, fmt.Errorf("ReadConfig: downloadpath not set")
	}

	if len(config.MusicPath) == 0 {
		log.Printf("\n\n-->It seems that you haven't already set a path for specifing which folder should be used to provide an offline song collection. Please set it at least to the same folder as the download directory! You can set this under: %s\n\n", configPath)
		return nil, fmt.Errorf("ReadConfig: path for music not set")
	}

	return config, nil
}
