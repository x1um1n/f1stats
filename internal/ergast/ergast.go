package ergast

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/x1um1n/checkerr"
	"github.com/x1um1n/f1stats/internal/shared"
	// "github.com/gomodule/redigo/redis"
)

// Constructor holds the info about constructors
type Constructor struct {
	ConstructorID      string   `json:"constructorId"`
	URL                string   `json:"url"`
	Name               string   `json:"name"`
	Nationality        string   `json:"nationality"`
	ConTitleCount      int      `json:"constructors-titles-count"`
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

// getChampConstructors gets all the Constructors who have won the constructors
// championship, the default limit is 30 and current count of unique Constructors
// is 17, so there should be no need to either get a second page or increase the
// results limit for the foreseeable future
func getChampConstructors() []Constructor {
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

// getConstructorsTitles gets all the years a constructor won the contructors
// championship.  as with getChampConstructors, there is no constructor which
// has won the title anywhere near 30 times
func getConstructorsTitles(con string) (titles []string) {
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

// getRaceStarts gets the total number of race starts for a constructor
func getRaceStarts(con string) int {
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

// getRaceWins gets the total number of race wins for a contructor
func getRaceWins(con string) int {
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

// getActiveYears gets all the seasons a constructor was active, currently
// limited to 70, will need increasing at some point
func getActiveYears(con string) (result []string) {
	log.Printf("Getting all seasons %s competed in\n", con)
	response, err := http.Get("https://ergast.com/api/f1/constructors/" + con + "/constructorStandings.json?limit=70")
	if !checkerr.Check(err, "Failed to get all seasons for ", con) {
		data, _ := ioutil.ReadAll(response.Body)
		var res struct {
			MRData struct {
				StandingsTable struct {
					Years []StandList `json:"StandingsLists"`
				} `json:"StandingsTable"`
			} `json:"MRData"`
		}
		json.Unmarshal(data, &res)

		for _, t := range res.MRData.StandingsTable.Years {
			result = append(result, t.Year)
		}
	}
	return
}

// getYearSpans takes a slice of strings, each containing a YYYY year and
// converting it to a span of years eg 1991-2000
func getYearSpans(years []string) string {
	type span struct {
		start string
		end   string
	}
	var spans []span
	spndx := 0

	for i, y := range years {
		if i == 0 {
			spans = []span{{y, y}}
			continue
		}
		st, _ := strconv.Atoi(y)
		nd, _ := strconv.Atoi(years[i-1])
		if st == nd+1 {
			spans[spndx].end = y
			continue
		} else { //if the years aren't consecutive, start a new span
			spndx++
			spans = append(spans, span{y, y})
			continue
		}
	}

	s := spans[0].start + "-" + spans[0].end
	for spndx = 1; spndx < len(spans); spndx++ {
		s += ", " + spans[spndx].start + "-" + spans[spndx].end
	}
	return s
}

// Repopulate empties the redis cache and get fresh stats from ergast
func Repopulate() error {
	c := shared.P.Get()
	defer c.Close()

	log.Println("Getting the latest f1 stats from ergast api")
	_, err := c.Do("FLUSHALL")
	if !checkerr.Check(err, "Error flushing redis, abandoning attempt to repopulate the data") {
		var teams []Constructor
		teams = getChampConstructors()

		for i, t := range teams {
			teams[i].ConstructorsTitles = getConstructorsTitles(t.ConstructorID)
			teams[i].ConTitleCount = len(teams[i].ConstructorsTitles)
			teams[i].RaceStarts = getRaceStarts(t.ConstructorID)
			teams[i].RaceWins = getRaceWins(t.ConstructorID)
			teams[i].WinRate = float32(teams[i].RaceWins) / float32(teams[i].RaceStarts)
			teams[i].WinRateH = fmt.Sprintf("%.2f%% (%d wins from %d starts)", (teams[i].WinRate * 100), teams[i].RaceWins, teams[i].RaceStarts)
			teams[i].YearsActive = getActiveYears(t.ConstructorID)
			teams[i].YearsActiveH = getYearSpans(teams[i].YearsActive)

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
