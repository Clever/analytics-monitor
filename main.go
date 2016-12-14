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

	fastConnection, err := db.NewRedshiftFastClient()
	if err != nil {
		log.Fatalln(err)
	}

	prodConnection, err := db.NewRedshiftProdClient()
	if err != nil {
		log.Fatalln(err)
	}

	// For testing TODO: remove once finished
	testConnections(fastConnection, "timeline.events")
	testConnections(prodConnection, "mongo.oauthclients")
}

// For testing TODO: remove once finished
func testConnections(redshiftClient *db.RedshiftClient, tableName string) {
	count, err := redshiftClient.CountRows(tableName)
	if err != nil {
		fmt.Printf("Error with client querying table %s: %v.\n", tableName, err)
	} else {
		fmt.Printf("Redshift has %d rows in %s\n", count, tableName)
	}
}
