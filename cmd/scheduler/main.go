// artist_first project main.go
package main

import (
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
	DBSTARTUP_TIMEOUT_SECS     = 30
	DELAY_BETWEEN_REPORTS_SECS = 60 * 60 // one hour
)

//
// scheduler.main()
// Represents the process body of the scheduler pod. Note that only one scheduler pod is required per install.
// The scheduler is responsible for creating new scanning batches and pushing the set of required scanning requests
// for the batch onto the scanning queue.
// @see scanner.main()
//
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

	// make sure that the db is up. keep trying every second until it is
	log.Debugf("Retail vinyl scanner starts..")
	err = vinylDS.WaitForDbUp(DBSTARTUP_TIMEOUT_SECS)
	if err != nil {
		panic(err)
	}

	for {

		// grab the list of known retailers
		retailers, err := vinylDS.GetAllRetailers(nil)
		if err != nil {
			log.Error(err, "Failed to get retailers list")
		}

		// grab the list of artists watched by which users
		watchedArtists, err := vinylDS.GetWatchedArtists(nil)
		if err != nil {
			log.Error(err, "Failed to get artists list")
		}

		// index a single scannable list of artists
		artists := map[int64]string{}
		for _, watches := range watchedArtists {
			for _, watch := range watches {
				artists[watch.ArtistID] = watch.ArtistName
			}
		}

		//
		// Create a new batch and enqueue all the scan requests for that batch.
		//

		var batchID int64
		if len(artists) > 0 && len(retailers) > 0 {
			// calculate the number of expected scans for a new batch
			requiredSearches := len(artists) * len(retailers)

			// add the new batch to the db (and start a report for each watching user)
			batchID, err = vinylDS.AddNewBatch(nil, requiredSearches, watchedArtists)
			if err != nil {
				log.Error(err, "Failed to start new batch")
			} else {
				// euqueue all our scan requests for the batch
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
							break
						}
					}
					if err != nil {
						break
					}
				}
				if err != nil {
					err2 := vinylDS.DeleteBatch(nil, batchID)
					if err2 != nil {
						log.Error(err, "Failed to cleanup batch %v", batchID)
					}
				}
			}
			if err != nil {
				log.Debugf("Failed to schedule new batch %v", batchID)
			} else {
				log.Debugf("Batch %v scheduled..", batchID)
			}
		}
		log.Debugf("Sleeping for %v seconds..", DELAY_BETWEEN_REPORTS_SECS)
		time.Sleep(time.Duration(DELAY_BETWEEN_REPORTS_SECS) * time.Second)
	}
}
