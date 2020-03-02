//Package server handles the queueing website
package server

import (
	"bufio"
	"fmt"
	"goparty/clients"
	"goparty/mp3"
	"goparty/youtube"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	templates        = template.Must(template.ParseFiles("html/user.html", "html/admin.html", "html/error.html", "html/songdb.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("(https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*|https{0,1}://youtu\\.be/\\S*)")
	serverIP         string
	config           *Config

	//configPath for unix systems
	configPath = ".config" + string(os.PathSeparator) + "goparty" + string(os.PathSeparator) + "config.json"
)

type userIP string

type errorUI struct {
	ErrorMsg string
}

type queueUI struct {
	Name    string
	IP      string
	AdminIP string
	Songs   []mp3.Song
}

func (ui queueUI) IsSongUpvotedByUser(songID int) bool {
	for _, elem := range ui.Songs[songID].GetUpvotes() {
		if elem == ui.IP {
			return true
		}
	}
	return false
}

type songdbUI struct {
	Songs []string
}

func (db songdbUI) IncreaseID(id int) int {
	return id + 1
}

func (db songdbUI) IsDirectory(path string) bool {
	if len(path) > 0 {
		if path[len(path)-1] == os.PathSeparator {
			return true
		}
	}
	return false
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	var uidata queueUI
	var ip userIP
	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(r.RemoteAddr, "::1") {
		ip = "127.0.0.1"
	} else {
		ip = userIP(r.RemoteAddr)
		ip = ip.normalizeIP()
	}

	uidata.Name = clients.GetUserNameToIP(ip.String())
	uidata.Songs = mp3.GetCurrentPlaylist()
	uidata.IP = ip.String()
	uidata.AdminIP = serverIP

	if i := r.FormValue("task"); len(i) != 0 {
		handleAdminTasks(i)
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if r.Method == "GET" {
		if strings.Contains(ip.String(), "127.0.0.1") || config.AllUserAdmin {
			renderTemplate(w, "admin", uidata)
		} else {
			renderTemplate(w, "user", uidata)
		}

	} else if r.Method == "POST" {
		link := r.FormValue("ytlink")
		if validYoutubeLink.MatchString(link) {
			fmt.Println("added: " + link)
			youtube.Add(link, ip.String())

			r.Method = "GET"
			http.Redirect(w, r, "/", http.StatusFound)
		} else if len(link) == 0 {
			http.Error(w, "500 - Unvalid POST Request", http.StatusInternalServerError)
		} else {
			renderTemplate(w, "error", errorUI{ErrorMsg: "You entered an unvalid Youtube-Link!"})
		}
	} else {
		fmt.Fprintf(w, "Only GET and POST methods are supported!")
	}
}

func handleAdminTasks(task string) {
	switch task {
	case "start":
		mp3.StartSpeaker()
	case "stop":
		mp3.CloseSpeaker()
	case "skip":
		mp3.SkipSong()
	case "pause":
		mp3.PauseSpeaker()
	default:
		log.Println("Unknown admin task received!")

	}
}

func upvoteHandler(w http.ResponseWriter, r *http.Request) {
	var ip userIP
	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(r.RemoteAddr, "::1") {
		ip = "127.0.0.1"
	} else {
		ip = userIP(r.RemoteAddr)
		ip = ip.normalizeIP()
	}

	if r.Method == "POST" {
		idStr := r.FormValue("id")

		id, err := strconv.Atoi(idStr)

		if err != nil {
			fmt.Printf("upvoteHandler: could not convert upvoted song id to int %s", err)
		}
		mp3.UpvoteSong(id, ip.String())
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		fmt.Fprintf(w, "Only POST methods are supported for /upvote!")
	}
}

func songDBHandler(w http.ResponseWriter, r *http.Request) {
	var ip userIP
	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(r.RemoteAddr, "::1") {
		ip = "127.0.0.1"
	} else {
		ip = userIP(r.RemoteAddr)
		ip = ip.normalizeIP()
	}

	if r.Method == "GET" {
		var dbui songdbUI
		dbui.Songs = mp3.GetSortedSongList()
		renderTemplate(w, "songdb", dbui)

	} else if r.Method == "POST" {

		songname := r.FormValue("offlineSongBtn")

		if mp3.CheckSongInDB(songname) {
			fmt.Println("Song exists!")
			filedir, complSongname := mp3.GetSongDirAndCompleteName(songname)

			err := mp3.AddMP3ToMusicQueue(filedir, complSongname, ip.String(), false)

			if err != nil {
				renderTemplate(w, "error", errorUI{ErrorMsg: "Could not add offline song: " + err.Error()})
				return
			}

			r.Method = "GET"
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			renderTemplate(w, "error", errorUI{ErrorMsg: "Could not find song in SongDB :("})
			return
		}
	}
}

//SetupServing sets up all we need to handle our "website"
func SetupServing() {

	homedir, err := os.UserHomeDir()

	if err != nil {
		log.Fatalln("SetupServing: ", err)
	}

	configPath = homedir + string(os.PathSeparator) + configPath

	cfg, err := ReadConfig(configPath)

	if err != nil {
		log.Fatalln(err)
	}

	config = cfg

	mp3.InitializeSongDBFromMemory(config.MusicPath, config.DownloadPath)

	err = clients.InitUserNames("usernames.txt")
	if err != nil {
		log.Println(err)
	}

	serverIP = getLocalServerAdress()

	//check for youtube-dl binary in $PATH
	youtube.MustExistYoutubeDL()

	setupMusic()

	mp3.SetNeededUpvoteCount(config.UpvotesNeededForRanking)

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)
	serverMux.HandleFunc("/upvote", upvoteHandler)
	serverMux.HandleFunc("/songdb", songDBHandler)

	youtube.StartDownloadWorker(config.DownloadPath, mp3.AddMP3ToMusicQueue)

	fmt.Println(createWelcomeMessage(serverIP))
	go handleUserInput()

	log.Fatal(http.ListenAndServe(":8080", serverMux))
}

func getLocalServerAdress() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		//shutdown the program at this stage, because something with the network card must be wrong
		log.Fatalln("Could not retrieve local music server IP address!", err)
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Fatalln("Could not retrieve local music server IP address!", err)
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			var ipv4 userIP = userIP(ip.String())

			if ipv4.String() != "127.0.0.1" {
				if ipv4.isIPv4() {
					return ipv4.String()
				}
			}
		}
	}
	return ""
}

func setupMusic() {
	err := mp3.InitSpeaker()
	if err != nil {
		log.Fatalln(err.Error())
	}
	mp3.StartSpeaker()
}

func (ip userIP) String() string {
	return string(ip)
}

func (ip userIP) normalizeIP() userIP {
	if strings.Contains(ip.String(), ":") {
		return userIP(strings.Split(ip.String(), ":")[0])
	}
	return ip
}

func (ip userIP) isIPv4() bool {
	if ip.String() != "127.0.0.1" {
		ip4Split := strings.Split(ip.String(), ".")
		if len(ip4Split) == 4 {
			return true
		}
		return false
	}
	return true
}

func createWelcomeMessage(ip string) string {
	builder := strings.Builder{}
	builder.WriteString("\n-------------------------------------\n")
	builder.WriteString("GOPARTY - The Youtube Music Queue\n")
	builder.WriteString("-------------------------------------\n\n")
	builder.WriteString("Hello you are the admin!\n")
	builder.WriteString("\033[0;31mYour local IP is: ")
	builder.WriteString(ip + "\033[0m\n")
	builder.WriteString("The config can be found under: ")
	builder.WriteString(configPath + "\n")
	builder.WriteString("After editing the config file you need to restart the program!")
	builder.WriteString("\n\n")
	builder.WriteString("Please open your webbrowser on this machine and enter: 'localhost:8080' for viewing the admin page!\n")
	builder.WriteString("All other users can view the website under: 'IP:8080'. The IP is written above.\n\n")
	builder.WriteString("You can enter the following commands:\n")
	builder.WriteString("- help (shows this text)\n")
	builder.WriteString("- play (starts paused music)\n")
	builder.WriteString("- pause (pauses the music)\n")
	builder.WriteString("- skip (skips the current playing song)\n")
	builder.WriteString("- list (lists all current songs in the playing queue)\n")
	builder.WriteString("- exit/quit (quits the program)\n")
	return builder.String()
}

func handleUserInput() error {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		ok := scanner.Scan()
		switch scanner.Text() {
		case "play":
			mp3.StartSpeaker()
		case "pause":
			mp3.PauseSpeaker()
		case "skip":
			mp3.SkipSong()
		case "list":
			fmt.Println()
			for _, song := range mp3.GetCurrentPlaylist() {
				fmt.Println(" - " + song.SongName + "\tby: " + song.UserName + " (" + song.UserIP + ")\tupvotes: " + strconv.Itoa(song.GetUpvotesCount()))
			}
			fmt.Println()
		case "help":
			fmt.Println(createWelcomeMessage(serverIP))
		case "exit", "quit", "q":
			fmt.Println("Quitting the program...")
			os.Exit(0)
		default:
			if len(scanner.Text()) > 0 {
				fmt.Println("unknown command!")
			}
		}

		if scanner.Err() != nil {
			return fmt.Errorf("handleUserInput: %s", scanner.Err().Error())
		}

		if ok == false {
			log.Println("Stopping listening to user input")
			break
		}
	}
	return nil
}
