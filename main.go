// artist_first project main.go
package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/util/cfg"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/postgres"
	_ "github.com/lib/pq"
	"os"
)

const (
	CITADEL_PREFIX = "https://www.citadelmailorder.com"
	CITADEL_URL    = "https://www.citadelmailorder.com/jsp/cmos/merch/hard_ons_merch_page.jsp?source=hard_ons_merch_page_new"
	PC_PREFIX      = "https://poisoncityestore.com"
	PC_URL         = "https://poisoncityestore.com/search?q=%s+vinyl"
	PERIOD_MINS    = 15
)

const (
	MAX_IDLES_CONNS = 2
)

func initialiseDbConnection() postgres.DB {
	maxConns, _ := cfg.IntSetting("DB_MAXCONNS")
	psqlDB, err := postgres.NewPostgresDB(postgres.MustGetOptsWithMaxConns(maxConns, MAX_IDLES_CONNS))
	if err != nil {
		log.Panic(err, "failed to connect to the postgres db")
		os.Exit(1)
	}
	return psqlDB
}

func main() {

	// initialise config (env vars, config file)
	cfg.InitConfig()

	// use the database config and initialise a postgres connection (will panic if incomplete)
	psqlDB := initialiseDbConnection()
	defer psqlDB.Close()

	log.Info("initialised psqlDB")
	fmt.Println("Vinyl checker")
}
