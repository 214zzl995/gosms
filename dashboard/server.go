package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/haxpax/gosms"
	"github.com/satori/go.uuid"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

// SMSResponse repose structure to /sms
type SMSResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// SMSDataResponse response structure to /medata/
type SMSDataResponse struct {
	Status   int            `json:"status"`
	Message  string         `json:"message"`
	Summary  []int          `json:"summary"`
	DayCount map[string]int `json:"day count"`
	Messages []gosms.SMS    `json:"messages"`
}

// Cache templates
var templates = template.Must(template.ParseFiles("./templates/index.html"))

var authUsername string
var authPassword string

/* dashboard handlers */

// dashboard
func indexHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("--- indexHandler")
	// templates.ExecuteTemplate(w, "index.html", nil)
	// Use during development to avoid having to restart server
	// after every change in HTML
	t, _ := template.ParseFiles("./templates/index.html")
	err := t.Execute(w, nil)
	if err != nil {
		return
	}
}

// handle all static files based on specified path
// for now its /assets
func handleStatic(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	static := vars["path"]
	http.ServeFile(w, r, filepath.Join("./assets", static))
}

/* end dashboard handlers */

/* API handlers */

// push sms, allowed methods: POST
func sendSMSHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("--- sendSMSHandler")
	w.Header().Set("Content-type", "application/json")

	//TODO: validation
	err := r.ParseForm()
	if err != nil {
		return
	}
	mobile := r.FormValue("mobile")
	message := r.FormValue("message")
	id := uuid.NewV4()

	sms := &gosms.SMS{UUID: id.String(), Mobile: mobile, Body: message, Retries: 0}
	gosms.EnqueueMessage(sms, true)

	smsResponse := SMSResponse{Status: 200, Message: "ok"}
	var toWrite []byte
	toWrite, err = json.Marshal(smsResponse)
	if err != nil {
		log.Println(err)
		//let's just depend on the server to raise 500
	}
	_, err = w.Write(toWrite)
	if err != nil {
		return
	}
}

// dumps JSON data, used by log view. Methods allowed: GET
func getLogsHandler(w http.ResponseWriter, _ *http.Request) {
	log.Println("--- getLogsHandler")
	messages, _ := gosms.GetMessages("")
	summary, _ := gosms.GetStatusSummary()
	dayCount, _ := gosms.GetLast7DaysMessageCount()
	logs := SMSDataResponse{
		Status:   200,
		Message:  "ok",
		Summary:  summary,
		DayCount: dayCount,
		Messages: messages,
	}
	var toWrite []byte
	toWrite, err := json.Marshal(logs)
	if err != nil {
		log.Println(err)
		//let's just depend on the server to raise 500
	}
	w.Header().Set("Content-type", "application/json")
	_, err = w.Write(toWrite)
	if err != nil {
		return
	}
}

/* end API handlers */

func InitServer(host string, port string, username string, password string) error {
	log.Println("--- InitServer ", host, port)

	authUsername = username
	authPassword = password

	r := mux.NewRouter()
	r.StrictSlash(true)

	r.HandleFunc("/", use(indexHandler, basicAuth))

	// handle static files
	r.HandleFunc(`/assets/{path:[a-zA-Z0-9=\-\/\.\_]+}`, use(handleStatic, basicAuth))

	// all API handlers
	api := r.PathPrefix("/api").Subrouter()
	api.Methods("GET").Path("/logs/").HandlerFunc(use(getLogsHandler, basicAuth))
	api.Methods("POST").Path("/sms/").HandlerFunc(use(sendSMSHandler, basicAuth))

	http.Handle("/", r)

	bind := fmt.Sprintf("%s:%s", host, port)
	log.Println("listening on: ", bind)
	return http.ListenAndServe(bind, nil)

}

// See https://gist.github.com/elithrar/7600878#comment-955958 for how to extend it to suit simple http.Handler's
func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(authUsername) == 0 {
			h.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			http.Error(w, "Not authorized", 401)
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			http.Error(w, err.Error(), 401)
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 || pair[0] != authUsername || pair[1] != authPassword {
			http.Error(w, "Not authorized", 401)
			return
		}

		h.ServeHTTP(w, r)
	}
}
