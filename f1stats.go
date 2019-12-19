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
	"log"
	"net/http"
	"time"
)

//PageVariables struct defining variables visible to the html pages
type PageVariables struct {
	PageTitle    string
	Constructors []ergast.Constructor
}

// defines and starts the healthcheck
func startHealth() {
	h := healthcheck.NewHandler()

	log.Println("Adding redis cache check")
	h.AddReadinessCheck("redis", healthcheck.Async(healthcheck.TCPDialCheck(shared.K.String("redis_host")+":6379", 50*time.Millisecond), 10*time.Second))

	go http.ListenAndServe("0.0.0.0:9080", h)
}

// Start index page handler which renders the all time constructors standings
func indexPage(w http.ResponseWriter, r *http.Request) {
	Title := "All Time F1 Constructors Standings"

	pv := PageVariables{
		PageTitle:    Title,
	}

	pv.Constructors = getConstructors()
	log.Println(pv)

	t, err := template.ParseFiles("web/template/index.html")
	checkerr.Check(err, "Index template parsing error")

	err = t.Execute(w, pv)
	checkerr.Check(err, "Index template executing error")
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

func main() {
	shared.LoadKoanf()         //read in the config
	shared.InitRedis() 				//create a redis connection pool

	ergast.Repopulate() //get a fresh dataset & load it into redis
	go startHealth()        //start the healthcheck endpoints

	http.HandleFunc("/", indexPage) //handler for the root page

	log.Fatal(http.ListenAndServe(":80", nil))
}
