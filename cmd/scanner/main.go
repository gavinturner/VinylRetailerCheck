package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/gavinturner/vinylretailers/retailers"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
	"strings"
	"time"
)

func main() {

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

	fmt.Printf("Retail vinyl scanner starts..\n")
	for {
		num, err := scanningQueue.QueueLength()
		if err != nil {
			log.Error(err, "Failed to get queue length")
		}
		if num > 0 {
			fmt.Printf("There are %v requests pending\n", num)
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
					same, err := vinylDS.UpsertSKU(nil, &sku)
					if err != nil {
						log.Error(err, "Failed to upsert new price for '%s' ", release.Name)
						continue
					}
					if !same {
						fmt.Printf("%s@%s: Found new release state: %s = %s (%v)\n", payload.ArtistName, payload.RetailerName, release.Name, sku.Price, sku.ID)
						// TODO: add this finding to a report for any interested users
					} else {
						fmt.Printf("%s@%s: releases [%v, %s] has not changed.\n", payload.ArtistName, payload.RetailerName, releaseID, release.Name)
					}

				}
			}
			fmt.Printf("Done\n")
		}
		// sleep 10 seconds
		time.Sleep(1000 * time.Millisecond)
	}
	fmt.Printf("Retail Scanner terminating..\n")
}
