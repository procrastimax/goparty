//Package server handles the queueing website
package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"
	"yt-queue/mp3"
	"yt-queue/youtube"
)

var (
	templates        = template.Must(template.ParseFiles("tmpl/addsong.html", "tmpl/admin.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*")

	downloadQueue youtube.DownloadQueue
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if strings.Contains(r.RemoteAddr, "127.0.0.1") || strings.Contains(r.RemoteAddr, "::1") {
			renderTemplate(w, "admin")
		} else {
			renderTemplate(w, "addsong")
		}

	} else if r.Method == "POST" {
		if strings.Contains(r.RemoteAddr, "127.0.0.1") || strings.Contains(r.RemoteAddr, "::1") {
			renderTemplate(w, "admin")
			fmt.Println(r.FormValue("startBtn"))
		} else {
			renderTemplate(w, "addsong")
		}
		link := r.FormValue("ytlink")
		if len(link) > 10 && validYoutubeLink.MatchString(link) {
			fmt.Println("Added:", link)
			downloadQueue.Add(link)
		} else {
			fmt.Fprintf(w, "\nYou entered a non-valid YoutTube link! Shame on you.")
		}
	} else {
		fmt.Fprintf(w, "Only GET and POST methods are supported!")
	}
}

func makeMusicHandler(fn func()) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/", http.StatusFound)
		fn()
	}
}

//SetupServing sets up all we need to handle our "website"
func SetupServing() {
	//check for youtube-dl binary in $PATH
	youtube.MustExistYoutubeDL()

	setupMusic()

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)
	serverMux.HandleFunc("/start", makeMusicHandler(mp3.StartSpeaker))
	serverMux.HandleFunc("/pause", makeMusicHandler(mp3.PauseSpeaker))
	serverMux.HandleFunc("/skip", makeMusicHandler(mp3.SkipSong))
	serverMux.HandleFunc("/stop", makeMusicHandler(mp3.CloseSpeaker))

	downloadQueue.StartDownloadWorker()

	log.Fatal(http.ListenAndServe(":8080", serverMux))
}

func setupMusic() {
	err := mp3.InitSpeaker()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
