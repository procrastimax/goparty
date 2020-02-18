//Package server handles the queueing website
package server

import (
	"fmt"
	"goparty/mp3"
	"goparty/user"
	"goparty/youtube"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	templates        = template.Must(template.ParseFiles("html/user.html", "html/admin.html", "html/error.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("(https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*|https{0,1}://youtu\\.be/\\S*)")
	uidata           uiData
)

type errorMessage struct {
	ErrorMsg string
}

type uiData struct {
	Name    string
	AdminIP string
	Songs   []mp3.Song
}

//returns a given ID increased by one, so we get 1 instead of 0 -> only for visual purpose
func (ui uiData) GetRealID(id int) int {
	return id + 1
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	userIP := r.RemoteAddr

	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(userIP, "::1") {
		userIP = "127.0.0.1:1234"
	}

	uidata.Name = user.GetUserNameToIP(userIP)
	uidata.Songs = mp3.GetCurrentPlaylist()

	if i := r.FormValue("task"); len(i) != 0 {
		handleAdminTasks(i)
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if r.Method == "GET" {
		if strings.Contains(userIP, "127.0.0.1") {
			renderTemplate(w, "admin", uidata)
		} else {
			renderTemplate(w, "user", uidata)
		}

	} else if r.Method == "POST" {
		link := r.FormValue("ytlink")
		if validYoutubeLink.MatchString(link) {
			youtube.Add(link, userIP)
			//we need to wait here shortly, so the website can update
			time.Sleep(150 * time.Millisecond)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			renderTemplate(w, "error", errorMessage{ErrorMsg: "You entered an unvalid Youtube-Link!"})
		}
	} else {
		fmt.Fprintf(w, "Only GET and POST methods are supported!")
	}
}

func upvoteHandler(w http.ResponseWriter, r *http.Request) {
	userIP := r.RemoteAddr

	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(userIP, "::1") {
		userIP = "127.0.0.1:1234"
	}

	if r.Method == "POST" {
		link := r.FormValue("id")
		fmt.Println(link)
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		fmt.Fprintf(w, "Only POST methods are supported!")
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

//SetupServing sets up all we need to handle our "website"
func SetupServing() {

	err := user.InitUserNames("usernames.txt")
	if err != nil {
		log.Println(err)
	}

	serverIP := getLocalServerAdress()

	uidata.AdminIP = serverIP

	//check for youtube-dl binary in $PATH
	youtube.MustExistYoutubeDL()

	//setupMusic()

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)
	serverMux.HandleFunc("/upvote", upvoteHandler)

	youtube.StartDownloadWorker(mp3.AddMP3ToMusicQueue)

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

			if ip.String() != "127.0.0.1" {
				ip4Split := strings.Split(ip.String(), ".")
				if len(ip4Split) == 4 {
					return ip.String()
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
