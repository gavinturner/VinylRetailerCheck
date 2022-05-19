package cmd

import (
	"github.com/gavinturner/vinylretailers/util/cfg"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/gavinturner/vinylretailers/util/redis"
	"github.com/pkg/errors"
)

const (
	DEFAULT_MAX_IDLE_CONNS = 2
	DEFAULT_MAX_CONNS      = 10
	REDIS_QUEUE_NAME       = "retailer_scanning_queue"
)

func init() {
	// initialise config (env vars, config file)
	cfg.InitConfig()
}

func InitialiseDbConnection() (*postgres.DB, error) {
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

func InitialiseRedisScanningQueue() (*redis.RedisQueue, error) {
	redisServer, _ := cfg.StringSetting("REDIS_SERVER")
	if redisServer == "" {
		redisServer = "localhost:6379"
	}
	redisPassword, _ := cfg.StringSetting("REDIS_PASSWORD")
	return redis.ConnectToQueue(redisServer, redisPassword, REDIS_QUEUE_NAME, true)
}
