package main

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/util/cfg"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/gavinturner/vinylretailers/util/redis"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"time"
)

const (
	PERIOD_MINS = 15
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

	// use the redis config and initialise a connection to the redis queue
	scanningQueue, err := initialiseRedisScanningQueue()
	if err != nil {
		panic(err)
	}
	defer scanningQueue.Close()
	_, err = scanningQueue.PingRedis()
	if err != nil {
		panic(err)
	}

	scanningQueue.Enqueue("payload_1")
	scanningQueue.Enqueue("payload_2")
	scanningQueue.Enqueue("payload_3")

	for {
		len, err := scanningQueue.QueueLength()
		if err != nil {
			fmt.Printf("Error getting queue length: %s\n", err.Error())
		}
		if len > 0 {
			fmt.Printf("There are %v requests pending\n", len)
			for {
				// check for requests on the scan queue and action until there are none
				payload := ""
				found, err := scanningQueue.Dequeue(&payload)
				if err != nil {
					fmt.Printf("Error dequeuing request: %s\n", err.Error())
					break
				} else if !found {
					break
				}
				// process payload here...
				fmt.Printf("REQUEST PAYLOAD: %s\n", payload)
			}
			fmt.Printf("Done\n")
		}
		// sleep 10 seconds
		time.Sleep(10000 * time.Millisecond)
	}
}
