package ergast

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/x1um1n/checkerr"
	// "github.com/x1um1n/f1stats/internal/shared"
	"github.com/gomodule/redigo/redis"
)

// Constructor holds the info about constructors
type Constructor struct {
	ConstructorID      string   `json:"constructorId"`
	URL                string   `json:"url"`
	Name               string   `json:"name"`
	Nationality        string   `json:"nationality"`
	ConstructorsTitles []string `json:"constructors-titles"`
	YearsActive        []string `json:"years-active"`       //fixme
	YearsActiveH       string   `json:"years-active-human"` //fixme
	RaceStarts         int      `json:"race-starts"`
	RaceWins           int      `json:"race-wins"`
	WinRate            float32  `json:"win-rate"`       //fixme
	WinRateH           string   `json:"win-rate-human"` //fixme
}

// StandList contains the years a Constructor has won the title
type StandList struct {
	Year string `json:"season"`
}

// GetChampConstructors gets all the Constructors who have won the constructors
// championship, the default limit is 30 and current count of unique Constructors
// is 17, so there should be no need to either get a second page or increase the
// results limit for the foreseeable future
func GetChampConstructors() []Constructor {
	log.Println("Getting all championship-winning constructors")
	response, err := http.Get("https://ergast.com/api/f1/constructorStandings/1/constructors.json")
	if !checkerr.Check(err, "Failed to get all championship-winning constructors") {
		data, _ := ioutil.ReadAll(response.Body)
		var res struct {
			ConsReslt struct {
				Limit   string `json:"limit"`
				Offset  string `json:"offset"`
				Total   string `json:"total"`
				ConsTab struct {
					Constructors []Constructor `json:"Constructors"`
				} `json:"ConstructorTable"`
			} `json:"MRData"`
		}
		json.Unmarshal(data, &res)
		return res.ConsReslt.ConsTab.Constructors
	}
	return nil
}

// GetConstructorsTitles gets all the years a constructor won the contructors
// championship.  as with getChampConstructors, there is no constructor which
// has won the title anywhere near 30 times
func GetConstructorsTitles(con string) (titles []string) {
	log.Printf("Getting all constructors titles for %s\n", con)
	response, err := http.Get("https://ergast.com/api/f1/constructors/" + con + "/constructorStandings/1.json")
	if !checkerr.Check(err, "Failed to get all constructors titles for ", con) {
		data, _ := ioutil.ReadAll(response.Body)
		var res struct {
			ConsReslt struct {
				Limit     string `json:"limit"`
				Offset    string `json:"offset"`
				Total     string `json:"total"`
				StandsTab struct {
					Years []StandList `json:"StandingsLists"`
				} `json:"StandingsTable"`
			} `json:"MRData"`
		}
		json.Unmarshal(data, &res)

		for _, t := range res.ConsReslt.StandsTab.Years {
			titles = append(titles, t.Year)
		}
		return
	}
	return nil
}

// GetRaceStarts gets the total number of race starts for a constructor
func GetRaceStarts(con string) int {
	log.Printf("Getting all race starts for %s\n", con)
	response, err := http.Get("https://ergast.com/api/f1/constructors/" + con + "/results.json?limit=0")
	if !checkerr.Check(err, "Failed to get all race starts for ", con) {
		data, _ := ioutil.ReadAll(response.Body)
		var res struct {
			MRData struct {
				Starts string `json:"total"`
			} `json:"MRData"`
		}
		json.Unmarshal(data, &res)

		s, e := strconv.Atoi(res.MRData.Starts)
		if !checkerr.Check(e, "Failed to convert string to int", res.MRData.Starts) {
			return s
		}
	}
	return 0
}

// GetRaceWins gets the total number of race wins for a contructor
func GetRaceWins(con string) int {
	log.Printf("Getting all race wins for %s\n", con)
	response, err := http.Get("https://ergast.com/api/f1/constructors/" + con + "/results/1.json?limit=0")
	if !checkerr.Check(err, "Failed to get all race wins for ", con) {
		data, _ := ioutil.ReadAll(response.Body)
		var res struct {
			MRData struct {
				Wins string `json:"total"`
			} `json:"MRData"`
		}
		json.Unmarshal(data, &res)

		s, e := strconv.Atoi(res.MRData.Wins)
		if !checkerr.Check(e, "Failed to convert string to int", res.MRData.Wins) {
			return s
		}
	}
	return 0
}

// Repopulate empties the redis cache and get fresh stats from ergast
func Repopulate(p *redis.Pool) error {
	c := p.Get()
	defer c.Close()

	log.Println("Getting the latest f1 stats from ergast api")
	_, err := c.Do("FLUSHALL")
	if !checkerr.Check(err, "Error flushing redis, abandoning attempt to repopulate the data") {
		var teams []Constructor
		teams = GetChampConstructors()

		for i, t := range teams {
			teams[i].ConstructorsTitles = GetConstructorsTitles(t.ConstructorID)
			teams[i].RaceStarts = GetRaceStarts(t.ConstructorID)
			teams[i].RaceWins = GetRaceWins(t.ConstructorID)
			teams[i].WinRate = float32(teams[i].RaceWins) / float32(teams[i].RaceStarts)
			teams[i].WinRateH = fmt.Sprintf("%f %% (%d wins from %d starts)", teams[i].WinRate, teams[i].RaceWins, teams[i].RaceStarts)

			json, e := json.Marshal(teams[i])
			if !checkerr.Check(e, "Error marshalling json") {
				_, err = c.Do("SET", t.ConstructorID, json)
				if checkerr.Check(err, "Error writing to redis:", string(json)) {
					return err
				}
			}
		}
	}
	return nil
}
