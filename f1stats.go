// f1stats is a web app drawing info from ergast.com to create all-time league
// tables for drivers & constructors
package main

import (
	//	"bytes"
	"encoding/json"
	"github.com/x1um1n/checkerr"
	"io/ioutil"
	"log"
	"net/http"
	// "strconv"
)

// Constructor holds the info about constructors
type Constructor struct {
	ConstructorID      string `json:"constructorId"`
	URL                string `json:"url"`
	Name               string `json:"name"`
	Nationality        string `json:"nationality"`
	DriversTitles      []string
	ConstructorsTitles []string
}

// ConsRes result set for constructor api query
type ConsRes struct {
	ConsReslt struct {
		Limit   string `json:"limit"`
		Offset  string `json:"offset"`
		Total   string `json:"total"`
		ConsTab struct {
			Constructors []Constructor `json:"Constructors"`
		} `json:"ConstructorTable"`
	} `json:"MRData"`
}

// StandList contains the years a Constructor has won the title
type StandList struct {
	Year string `json:"season"`
}

// ConsWinsRes result set for Constructor titles api query
type ConsWinsRes struct {
	ConsReslt struct {
		Limit     string `json:"limit"`
		Offset    string `json:"offset"`
		Total     string `json:"total"`
		StandsTab struct {
			Years []StandList `json:"StandingsLists"`
		} `json:"StandingsTable"`
	} `json:"MRData"`
}

// getChampConstructors gets all the Constructors who have won the constructors
// championship, the default limit is 30 and current count of unique Constructors
// is 17, so there should be no need to either get a second page or increase the
// results limit for the foreseeable future
func getChampConstructors() []Constructor {
	log.Println("Getting all championship-winning constructors from ergast api")
	response, err := http.Get("https://ergast.com/api/f1/constructorStandings/1/constructors.json")
	if !checkerr.Check(err, "Failed to get all championship-winning constructors") {
		data, _ := ioutil.ReadAll(response.Body)
		var res ConsRes
		json.Unmarshal(data, &res)
		return res.ConsReslt.ConsTab.Constructors
	}
	return nil
}

// getConstructorsTitles gets all the years a constructor won the contructors
// championship.  as with getChampConstructors, there is no constructor which
// has won the title anywhere near 30 times
func getConstructorsTitles(con string) (titles []string) {
	log.Printf("Getting all constructors titles for %s from ergast api\n", con)
	response, err := http.Get("https://ergast.com/api/f1/constructors/" + con + "/constructorStandings/1.json")
	if !checkerr.Check(err, "Failed to get all constructors titles for ", con) {
		data, _ := ioutil.ReadAll(response.Body)
		var res ConsWinsRes
		json.Unmarshal(data, &res)

		for _, t := range res.ConsReslt.StandsTab.Years {
			titles = append(titles, t.Year)
		}
		return
	}
	return nil
}

func main() {
	var teams []Constructor
	teams = getChampConstructors()

	for i, t := range teams {
		teams[i].ConstructorsTitles = getConstructorsTitles(t.ConstructorID)
		log.Printf("%s won the constructors title %d times: ", t.Name, len(teams[i].ConstructorsTitles))
		for _, tt := range teams[i].ConstructorsTitles {
			log.Printf("%s ", tt)
		}
		log.Printf("\n")
	}
}
