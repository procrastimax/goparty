//Package server handles the queueing website
package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"yt-queue/mp3"
	"yt-queue/youtube"
)

var (
	templates        = template.Must(template.ParseFiles("tmpl/user.html", "tmpl/admin.html", "tmpl/error.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*")
	playlist         songList
)

type errorMessage struct {
	ErrorMsg string
}

type songList struct {
	Songs []mp3.Song
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	userIP := r.RemoteAddr
	playlist.Songs = mp3.GetCurrentPlaylist()
	if r.Method == "GET" {
		if strings.Contains(userIP, "127.0.0.1") || strings.Contains(userIP, "::1") {
			renderTemplate(w, "admin", playlist)
		} else {
			renderTemplate(w, "user", playlist)
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

func makeAdminHandler(fn func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		userIP := r.RemoteAddr
		if strings.Contains(userIP, "127.0.0.1") || strings.Contains(userIP, "::1") {
			http.Redirect(w, r, "/", http.StatusFound)
			fn()
		} else {
			renderTemplate(w, "error", errorMessage{ErrorMsg: "You don't have permissions to do this! Only the admin machine can do this!"})
		}
	}
}

//SetupServing sets up all we need to handle our "website"
func SetupServing() {
	//check for youtube-dl binary in $PATH
	youtube.MustExistYoutubeDL()

	setupMusic()

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)
	serverMux.HandleFunc("/start", makeAdminHandler(mp3.StartSpeaker))
	serverMux.HandleFunc("/pause", makeAdminHandler(mp3.PauseSpeaker))
	serverMux.HandleFunc("/skip", makeAdminHandler(mp3.SkipSong))
	serverMux.HandleFunc("/stop", makeAdminHandler(mp3.CloseSpeaker))
	serverMux.HandleFunc("/stopDL", makeAdminHandler(youtube.ExitDownloadWorker))

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
