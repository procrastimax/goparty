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
	"strconv"
	"strings"
	"time"
)

var (
	templates        = template.Must(template.ParseFiles("html/user.html", "html/admin.html", "html/error.html"))
	validPath        = regexp.MustCompile("^/(start|skip|pause|stop)")
	validYoutubeLink = regexp.MustCompile("(https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*|https{0,1}://youtu\\.be/\\S*)")
	uidata           uiData
)

type userIP string

type errorMessage struct {
	ErrorMsg string
}

type uiData struct {
	Name    string
	IP      string
	AdminIP string
	Songs   []mp3.Song
}

func (ui uiData) IsSongUpvotedByUser(songID int) bool {
	for _, elem := range ui.Songs[songID].GetUpvotes() {
		if elem == ui.IP {
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
	var ip userIP
	//convert localhost ipv6 resolution to an ipv4 address
	if strings.Contains(r.RemoteAddr, "::1") {
		ip = "127.0.0.1"
	} else {
		ip = userIP(r.RemoteAddr)
		ip = ip.normalizeIP()
	}

	uidata.Name = user.GetUserNameToIP(ip.String())
	uidata.Songs = mp3.GetCurrentPlaylist()
	uidata.IP = ip.String()

	if i := r.FormValue("task"); len(i) != 0 {
		handleAdminTasks(i)
		http.Redirect(w, r, "/", http.StatusFound)
	}

	if r.Method == "GET" {
		if strings.Contains(ip.String(), "127.0.0.1") {
			renderTemplate(w, "admin", uidata)
		} else {
			renderTemplate(w, "user", uidata)
		}

	} else if r.Method == "POST" {
		link := r.FormValue("ytlink")
		if validYoutubeLink.MatchString(link) {
			youtube.Add(link, ip.String())
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

	setupMusic()

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
