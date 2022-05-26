package main

import (
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/gavinturner/vinylretailers/retailers"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

const (
	STARTUP_DELAY_SECS = 10
	SOLD_OUT           = "sold out"
)

func main() {

	// sleep on startup to let the infra pods get started up
	time.Sleep(time.Duration(STARTUP_DELAY_SECS) * time.Second)

	// use the database config and initialise a postgres connection (will panic if incomplete)
	psqlDB, err := cmd.InitialiseDbConnection()
	if err != nil {
		panic(err)
	}
	defer psqlDB.Close()
	vinylDS := db.NewDB(psqlDB)

	// use the redis config to initialise a connection to the redis scanning queue
	scanningQueue, err := cmd.InitialiseRedisScanningQueue()
	if err != nil {
		panic(err)
	}
	defer scanningQueue.Close()
	_, err = scanningQueue.PingRedis()
	if err != nil {
		panic(err)
	}

	log.Debugf("Retail vinyl scanner starts..")

	err = vinylDS.VerifySchema()
	if err != nil {
		log.Error(err, "Database is not valid?")
	}
	log.Debugf("Database looks ok..\n")

	for {
		num, err := scanningQueue.QueueLength()
		if err != nil {
			log.Error(err, "Failed to get queue length")
		}
		if num > 0 {
			log.Debugf("There are %v requests pending..", num)
			for {
				// check for requests on the scan queue and action until there are none
				payload := redis.ScanRequest{}
				found, err := scanningQueue.Dequeue(&payload)
				if err != nil {
					log.Error(err, "Failed to dequeue request")
					break
				} else if !found {
					break
				}
				// get the scraper implementaiton for the nominated retailer
				retailerScraper, err := retailers.VinylRetailerFactory(retailers.RetailerID(payload.RetailerID))
				if err != nil {
					log.Error(err, "Could not determine scraper for retailer (%v) %s", payload.RetailerID, payload.RetailerName)
				}
				// scrape for available releases
				releases, err := retailerScraper.ScrapeArtistReleases(strings.TrimSpace(strings.ToLower(payload.ArtistName)))
				if err != nil {
					log.Error(err, "Failed to scrape '%s' for '%s'", payload.RetailerName, payload.ArtistName)
				}
				for _, release := range releases {
					
					releaseID, err := vinylDS.UpsertRelease(nil, payload.ArtistID, release.Name)
					if err != nil {
						log.Error(err, "Failed to create new release '%s' for artist '%s'", release.Name, payload.ArtistName)
						continue
					}
					sku := db.SKU{
						ReleaseID:  releaseID,
						RetailerID: payload.RetailerID,
						ArtistID:   payload.ArtistID,
						ItemUrl:    release.Url,
						ImageUrl:   release.Image,
						Price:      release.Price,
					}
					err = vinylDS.IncrementBatchSearchCompletedCount(nil, payload.BatchID)
					if err != nil {
						log.Error(err, "Failed to increment search count for batch %v", payload.BatchID)
						continue
					}
					same, err := vinylDS.UpsertSKU(nil, &sku)
					if err != nil {
						log.Error(err, "Failed to upsert new price for '%s' ", release.Name)
						continue
					}
					if !same && sku.Price != SOLD_OUT {
						log.Debugf("%s@%s: Found new release state: %s = %s (%v)", payload.ArtistName, payload.RetailerName, release.Name, sku.Price, sku.ID)
						err = vinylDS.AddSKUToReportsForBatch(nil, payload.BatchID, &sku)
						if err != nil {
							log.Error(err, "Failed to upsert new price for '%s' ", release.Name)
							continue
						}
					} else {
						log.Debugf("%s@%s: releases [%v, %s] has not changed", payload.ArtistName, payload.RetailerName, releaseID, release.Name)
					}
				}
				err = vinylDS.IncrementBatchSearchCompletedCount(nil, payload.BatchID)
				if err != nil {
					log.Error(err, "Failed to increment search count for batch %v ", payload.BatchID)
				}
			}
			log.Debugf("No more pending requests")
		}
		// sleep 10 seconds
		time.Sleep(1000 * time.Millisecond)
	}
	log.Debugf("Retail Scanner terminating..")
}
