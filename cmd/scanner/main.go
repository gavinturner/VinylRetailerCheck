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
	STARTUP_DELAY_SECS     = 10
	DBSTARTUP_TIMEOUT_SECS = 30
	SOLD_OUT               = "sold out"
)

//
// scanner.main()
// Represents the process body of the scanner pod. A single scanner pod essentially listens on the redis scanner
// queue waiting for new scan requests. Each scan request is for an artist + retailer and is for a specific batch.
// The scanner takes this request and runs the appropriate retail scraper for the nominated retailer, seeding with
// the nominated artist. Any new/updated releases found are stored as release SKUs in the database and attached to
// any open batch reports that include that artist (typically one report per watching user).
// @see scheduler.main()
//
func main() {

	// sleep on startup to let the infra pods get started
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

	// make sure that the db is up. keep trying every second until it is
	log.Debugf("Retail vinyl scanner starts..")
	err = vinylDS.WaitForDbUp(DBSTARTUP_TIMEOUT_SECS)
	if err != nil {
		panic(err)
	}

	//
	// check for requests on the scan queue (blocking) and action when found by calling the correct scraper
	// for the nominated retailer. Any SKUs found are compared with what we already know for the associated release
	// and the SKU is updated if we have new information (price / availability)
	//

	for {
		// grab the next request
		payload := redis.ScanRequest{}
		found, err := scanningQueue.Dequeue(&payload, true)
		if err != nil {
			log.Error(err, "Failed to dequeue request")
			break
		} else if !found {
			log.Error(err, "Failed to block on request dequeue?")
			break
		}
		// get the scraper implementation for the nominated retailer
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
			// upsert the release to create if we haven't seen it before. otherwise get release id
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
			// upsert a new SKU for the release. A new SKU record will be created if the price/availabilioty
			// of the release has changed (as compared to the most recent existing SKU for the release)
			same, err := vinylDS.UpsertSKU(nil, &sku)
			if err != nil {
				log.Error(err, "Failed to upsert new price for '%s' ", release.Name)
				continue
			}

			// if the sku is available and the price has changed, then it's a candidate for adding to one
			// or more user reports for the current batch
			if !same && sku.Price != SOLD_OUT {
				log.Debugf("%s@%s: Found new release state: %s = %s (%v)", payload.ArtistName, payload.RetailerName, release.Name, sku.Price, sku.ID)
				err = vinylDS.AddSKUToReportsForBatch(nil, payload.BatchID, &sku)
				if err != nil {
					log.Error(err, "Failed to upsert new price for '%s' ", release.Name)
					continue
				}
			} else if !same && sku.Price == SOLD_OUT {
				log.Debugf("%s@%s: Found new (sold out) release state: %s = %s (%v)", payload.ArtistName, payload.RetailerName, release.Name, sku.Price, sku.ID)
			} else {
				log.Debugf("%s@%s: releases [%v, %s] has not changed", payload.ArtistName, payload.RetailerName, releaseID, release.Name)
			}
		}

		// now that we've finished processing the artist + retailer, we can increment the number of required
		// searches for the current batch (so we know when the batch is done)
		err = vinylDS.IncrementBatchSearchCompletedCount(nil, payload.BatchID)
		if err != nil {
			log.Error(err, "Failed to increment search count for batch %v ", payload.BatchID)
		}
	}
	log.Debugf("Retail Scanner terminating..")
}
