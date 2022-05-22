// artist_first project main.go
package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
	"time"
)

const (
	// delay for ten minutes between attepts to produce reports
	STARTUP_DELAY_SECS         = 10
	DELAY_BETWEEN_REPORTS_SECS = 60 * 60 // one hour
)

func main() {

	// sleep on startup to let the infra pods get started up
	time.Sleep(time.Duration(STARTUP_DELAY_SECS) * time.Second)

	// use the database config and initialise a postgres connection (will panic if incomplete)
	psqlDB, err := cmd.InitialiseDbConnection()
	if err != nil {
		panic(err)
	}
	if psqlDB == nil {
		panic("db pointer is null?")
	}
	defer psqlDB.Close()
	vinylDS := db.NewDB(psqlDB)

	scanningQueue, err := cmd.InitialiseRedisScanningQueue()
	if err != nil {
		panic(err)
	}
	defer scanningQueue.Close()
	_, err = scanningQueue.PingRedis()
	if err != nil {
		panic(err)
	}

	log.Debugf("Retail vinyl scheduler starts..\n")

	err = vinylDS.VerifySchema()
	if err != nil {
		log.Error(err, "Database is not valid?")
	}
	log.Debugf("Database looks ok..")

	for {

		// grab the list of retailers
		retailers, err := vinylDS.GetAllRetailers(nil)
		if err != nil {
			log.Error(err, "Failed to get retailers list")
		}

		// grab the list of artists watched by users
		watchedArtists, err := vinylDS.GetWatchedArtists(nil)
		if err != nil {
			log.Error(err, "Failed to get artists list")
		}

		artists := map[int64]string{}
		for _, watches := range watchedArtists {
			for _, watch := range watches {
				artists[watch.ArtistID] = watch.ArtistName
			}
		}

		fmt.Printf("Watched Artists: %v\n", watchedArtists)

		var batchID int64
		if len(artists) > 0 && len(retailers) > 0 {
			requiredSearches := len(artists) * len(retailers)
			batchID, err = vinylDS.AddNewBatch(nil, requiredSearches, watchedArtists)
			if err != nil {
				log.Error(err, "Failed to start new batch")
			} else {
				log.Debugf("Scheduling batch %v...", batchID)
				for _, retailer := range retailers {
					for artistID, artistName := range artists {
						payload := redis.ScanRequest{
							BatchID:      batchID,
							ArtistID:     artistID,
							RetailerID:   retailer.ID,
							ArtistName:   artistName,
							RetailerName: retailer.Name,
						}
						err = scanningQueue.Enqueue(payload)
						if err != nil {
							log.Error(err, "Failed to write scanning request to redis scanning queue")
						}
					}
				}
			}
			log.Debugf("Batch %v scheduled..", batchID)
		}
		log.Debugf("Sleeping for %v seconds..", DELAY_BETWEEN_REPORTS_SECS)
		time.Sleep(time.Duration(DELAY_BETWEEN_REPORTS_SECS) * time.Second)
	}
}
