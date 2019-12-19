// f1stats is a web app drawing info from ergast.com to create all-time league
// tables for drivers & constructors
package main

import (
	// "github.com/x1um1n/checkerr"
	"github.com/x1um1n/f1stats/internal/ergast"
	"github.com/x1um1n/f1stats/internal/shared"
	"log"
)

func main() {
	conn := shared.InitRedis()

	var teams []ergast.Constructor
	teams = ergast.GetChampConstructors()

	for i, t := range teams {
		teams[i].ConstructorsTitles = ergast.GetConstructorsTitles(t.ConstructorID)
		log.Printf("%s won the constructors title %d times: ", t.Name, len(teams[i].ConstructorsTitles))
		for _, tt := range teams[i].ConstructorsTitles {
			log.Printf("%s ", tt)
		}
		log.Printf("\n")
	}
}
