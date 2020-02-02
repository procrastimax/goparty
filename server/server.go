//Package server handles the queueing website
package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var (
	templates        = template.Must(template.ParseFiles("tmpl/addsong.html"))
	validPath        = regexp.MustCompile("^/(add|skip|pause|resume|stop)/([a-zA-Z0-9]+)$")
	validYoutubeLink = regexp.MustCompile("https{0,1}://www\\.youtube\\.com/watch\\?v=\\S*")
)

func renderTemplate(w http.ResponseWriter, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl+".html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "addsong")
	} else if r.Method == "POST" {
		renderTemplate(w, "addsong")
		link := r.FormValue("ytlink")
		if len(link) > 10 && validYoutubeLink.MatchString(link) {
			fmt.Println(validYoutubeLink.FindString(link))
		} else {
			fmt.Fprintf(w, "\nYou entered a non-valid YoutTube link! Shame on you.")
		}
	} else {
		fmt.Fprintf(w, "Only GET and POST methods are supported!")
	}
}

func errorHandler(w http.ResponseWriter, r *http.Request) {

}

//SetupServing sets up all we need to handle our "website"
func SetupServing() {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", viewHandler)
	log.Fatal(http.ListenAndServe(":8080", serverMux))
}
