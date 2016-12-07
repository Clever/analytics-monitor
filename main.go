package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Clever/analytics-pipeline-monitor/config"
	"github.com/Clever/analytics-pipeline-monitor/db"
)

func main() {
	flag.Parse()
	config.Parse()

	testConnections()
}

// For testing
func testConnections() {
	fastConnection, err := db.NewRedshiftFastClient()
	if err != nil {
		log.Fatalln(err)
	}

	prodConnection, err := db.NewRedshiftProdClient()
	if err != nil {
		log.Fatalln(err)
	}

	count, err := fastConnection.CountRows("timeline.events")
	if err != nil {
		fmt.Printf("Error with fast prod query: %v.\n", err)
	} else {
		fmt.Printf("Fast prod has %d timeline events\n", count)
	}

	count, err = prodConnection.CountRows("mongo.oauthclients")
	if err != nil {
		fmt.Printf("Error with prod query: %v.\n", err)
	} else {
		fmt.Printf("Prod has %d mongo oauthclients \n", count)
	}
}
