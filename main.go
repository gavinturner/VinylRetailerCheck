// artist_first project main.go
package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/util/cfg"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

const (
	CITADEL_PREFIX = "https://www.citadelmailorder.com"
	CITADEL_URL    = "https://www.citadelmailorder.com/jsp/cmos/merch/hard_ons_merch_page.jsp?source=hard_ons_merch_page_new"
	PC_PREFIX      = "https://poisoncityestore.com"
	PC_URL         = "https://poisoncityestore.com/search?q=%s+vinyl"
	PERIOD_MINS    = 15
)

const (
	DEFAULT_MAX_IDLE_CONNS = 2
	DEFAULT_MAX_CONNS      = 10
	REDIS_QUEUE_NAME       = "retailer_scanning_queue"
)

func initialiseDbConnection() (*postgres.DB, error) {
	maxConns, _ := cfg.IntSetting("DB_MAXCONNS")
	if maxConns == 0 {
		maxConns = DEFAULT_MAX_CONNS
	}
	maxIdleConns, _ := cfg.IntSetting("DB_MAX_IDLE_CONNS")
	if maxIdleConns == 0 {
		maxIdleConns = DEFAULT_MAX_IDLE_CONNS
	}
	psqlDB, err := postgres.NewPostgresDB(postgres.MustGetOptsWithMaxConns(maxConns, maxIdleConns))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to the postgres db")
	}
	return &psqlDB, nil
}

func initialiseRedisScanningQueue() (*redis.RedisQueue, error) {
	redisServer, _ := cfg.StringSetting("REDIS_SERVER")
	if redisServer == "" {
		redisServer = "localhost:6379"
	}
	redisPassword, _ := cfg.StringSetting("REDIS_PASSWORD")
	return redis.ConnectToQueue(redisServer, redisPassword, REDIS_QUEUE_NAME, true)
}

func main() {

	// initialise config (env vars, config file)
	cfg.InitConfig()

	// use the database config and initialise a postgres connection (will panic if incomplete)
	psqlDB, err := initialiseDbConnection()
	if err != nil {
		panic(err)
	}
	defer psqlDB.Close()

	scanningQueue, err := initialiseRedisScanningQueue()
	if err != nil {
		panic(err)
	}
	defer scanningQueue.Close()

	_, err = scanningQueue.PingRedis()
	if err != nil {
		panic(err)
	}

	for i := 0; i < 100; i++ {
		err = scanningQueue.Enqueue(fmt.Sprintf("payload_%v", i))
	}

}
