//Package server handles the queueing website
package server

import (
	"fmt"
	"goparty/mp3"
	"goparty/user"
	"goparty/youtube"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var (
	templates        = template.Must(template.ParseFiles("html/user.html", "html/admin.html", "html/error.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*")
	uidata           uiData
)

type errorMessage struct {
	ErrorMsg string
}

type uiData struct {
	UserName string
	Songs    []string
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
		userIP = "127.0.0.1"
	}

	uidata.UserName = user.GetUserNameToIP(userIP)
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
			time.Sleep(100 * time.Millisecond)
			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			renderTemplate(w, "error", errorMessage{ErrorMsg: "You entered an unvalid Youtube-Link!"})
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

//SetupServing sets up all we need to handle our "website"
func SetupServing() {

	err := user.InitUserNames("usernames.txt")
	if err != nil {
		log.Println(err)
	}

	//check for youtube-dl binary in $PATH
	youtube.MustExistYoutubeDL()

	setupMusic()

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)

	youtube.StartDownloadWorker(mp3.AddMP3ToMusicQueue)

	log.Fatal(http.ListenAndServe(":8080", serverMux))
}

func setupMusic() {
	err := mp3.InitSpeaker()
	if err != nil {
		log.Fatalln(err.Error())
	}
	mp3.StartSpeaker()
}
