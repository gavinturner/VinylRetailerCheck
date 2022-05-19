// artist_first project main.go
package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/cmd"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/gavinturner/vinylretailers/util/cfg"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
)

const (
	CITADEL_PREFIX = "https://www.citadelmailorder.com"
	CITADEL_URL    = "https://www.citadelmailorder.com/jsp/cmos/merch/hard_ons_merch_page.jsp?source=hard_ons_merch_page_new"
	PC_PREFIX      = "https://poisoncityestore.com"
	PC_URL         = "https://poisoncityestore.com/search?q=%s+vinyl"
	PERIOD_MINS    = 15
)

func main() {

	// initialise config (env vars, config file)
	cfg.InitConfig()

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

	// grab the list of retailers
	retailers, err := vinylDS.GetAllRetailers(nil)
	if err != nil {
		panic(err)
	}

	// grab the list of artists
	artists, err := vinylDS.GetAllArtists(nil)
	if err != nil {
		panic(err)
	}

	for _, retailer := range retailers {
		for _, artist := range artists {
			payload := redis.ScanRequest{
				ArtistID:     artist.ID,
				RetailerID:   retailer.ID,
				ArtistName:   artist.Name,
				RetailerName: retailer.Name,
			}
			err = scanningQueue.Enqueue(payload)
			if err != nil {
				log.Error(err, "Failed to write scanning request to redis scanning queue")
			}
		}
	}
}

func renderResultRow(new bool, image string, artist string, url string, name string, price string, existingPrice string) string {
	htmlOut := "<tr>\n"
	htmlOut += fmt.Sprintf("<td>%s</td>\n", image)
	htmlOut += fmt.Sprintf("<td>%s<br>%s%s</a><br>%s</td>\n", artist, url, name, price)
	htmlOut += "</tr>\n"
	return htmlOut
}
