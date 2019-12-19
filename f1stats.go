// f1stats is a web app drawing info from ergast.com to create all-time league
// tables for drivers & constructors
package main

import (
	// "github.com/x1um1n/checkerr"
	"github.com/heptiolabs/healthcheck"
	"github.com/x1um1n/f1stats/internal/ergast"
	"github.com/x1um1n/f1stats/internal/shared"

	"log"
	"net/http"
	"time"
)

// defines and starts the healthcheck
func startHealth() {
	h := healthcheck.NewHandler()

	log.Println("Adding redis cache check")
	h.AddReadinessCheck("redis", healthcheck.Async(healthcheck.TCPDialCheck(shared.K.String("redis_host")+":6379", 50*time.Millisecond), 10*time.Second))

	go http.ListenAndServe("0.0.0.0:9080", h)
}

func main() {
	shared.LoadKoanf()         //read in the config
	pool := shared.InitRedis() //create a redis connection pool

	go startHealth() //start the healthcheck endpoints

	ergast.Repopulate(pool)

	//fixme: listen and serve
}
