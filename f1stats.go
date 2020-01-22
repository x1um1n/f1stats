// f1stats is a web app drawing info from ergast.com to create all-time league
// tables for drivers & constructors
package main

import (
	"github.com/gomodule/redigo/redis"
	"github.com/heptiolabs/healthcheck"
	"github.com/x1um1n/checkerr"
	"github.com/x1um1n/f1stats/internal/ergast"
	"github.com/x1um1n/f1stats/internal/shared"

	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

//PageVariables struct defining variables visible to the html pages
type PageVariables struct {
	PageTitle    string
	Constructors []ergast.Constructor
	CSSVer       string
}

//StartTime holds a timestamp of when the app started, for cache-busting
var StartTime string

// defines and starts the healthcheck
func startHealth() {
	h := healthcheck.NewHandler()

	log.Println("Adding redis cache check")
	h.AddReadinessCheck("redis", healthcheck.Async(healthcheck.TCPDialCheck(shared.K.String("redis_host")+":6379", 50*time.Millisecond), 10*time.Second))

	go http.ListenAndServe("0.0.0.0:9080", h)
}

func getConstructors() (result []ergast.Constructor) {
	c := shared.P.Get()
	defer c.Close()

	log.Println("Getting constructors stats from redis")
	teams, err := redis.Strings(c.Do("KEYS", "*"))
	if !checkerr.Check(err, "Error getting keys from redis") {
		for _, t := range teams {
			res := ergast.Constructor{}
			s, err2 := redis.String(c.Do("GET", t))
			if !checkerr.Check(err2, "Error getting team info for", t) {
				json.Unmarshal([]byte(s), &res)
			}
			result = append(result, res)
		}
	}
	return
}

// updateCalendar gets the latest ical file for the f1fanatics calendar from google
func updateCalendar() error {
	log.Println("Getting F1 calendar..")
	resp, err := http.Get("https://www.google.com/calendar/ical/hendnaic1pa2r3oj8b87m08afg%40group.calendar.google.com/public/basic.ics")
	checkerr.Check(err, "Error downloading calendar")
	defer resp.Body.Close()

	out, err := os.Create("assets/basic.ics")
	checkerr.Check(err, "Error creating local file")
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	checkerr.Check(err, "Error writing calendar to local storage")

	return err
}

/******************** HTTP handlers *******************************************/

// Start index page handler which renders the all time constructors standings
func indexPage(w http.ResponseWriter, r *http.Request) {
	Title := "All Time F1 Constructors Standings"

	pv := PageVariables{
		PageTitle: Title,
		CSSVer:    StartTime,
	}

	pv.Constructors = getConstructors()

	t, err := template.ParseFiles("web/template/index.html")
	checkerr.Check(err, "Index template parsing error")

	err = t.Execute(w, pv)
	checkerr.Check(err, "Index template executing error")
}

// repop is a handler for ergast.Repopu
func repop(w http.ResponseWriter, r *http.Request) {
	checkerr.Check(ergast.Repopulate(), "Failed to repopulate redis cache from ergast")
}

// refresh is a handler for ergast.RefreshRaceStats
func refresh(w http.ResponseWriter, r *http.Request) {
	checkerr.Check(ergast.RefreshRaceStats(), "Failed to repopulate redis cache from ergast")
}

/******************** End HTTP handlers ***************************************/

func main() {
	StartTime = time.Now().String() //capture the start time for cache-busting
	shared.LoadKoanf()              //read in the config
	shared.InitRedis()              //create a redis connection pool

	if updateCalendar() != nil {
		log.Println("Failed to get calendar, retrying..")
		if updateCalendar() != nil {
			log.Println("Failed to get calendar, again...fuck this shit..")
		}
	}

	go startHealth()                                                                                       //start the healthcheck endpoints
	http.HandleFunc("/", indexPage)                                                                        //handler for the root page
	http.Handle("/web/static/", http.StripPrefix("/web/static/", http.FileServer(http.Dir("web/static")))) //expose images & css
	http.HandleFunc("/repop", repop)                                                                       //get a fresh dataset & load it into redis
	http.HandleFunc("/refresh", refresh)                                                                   //get a fresh dataset & load it into redis

	log.Fatal(http.ListenAndServe(":80", nil))
}
