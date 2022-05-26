package redis

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"time"
)

const (
	BLOCKING_DELAY_MSECS = 500
	MASTER_QUEUE_KEY     = "vinylretailers::queue"
)

type RedisQueue struct {
	client *redis.Client
	name   string
}

func ConnectToQueue(redisServer string, redisPassword string, queueName string, create bool) (*RedisQueue, error) {
	r := &RedisQueue{
		client: redis.NewClient(&redis.Options{
			Addr:     redisServer,
			Password: redisPassword,
			DB:       0, // default
		}),
		name: queueName,
	}
	exists, err := r.PingRedis()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to contact redis server at %v", redisServer)
	}
	if !exists {
		return nil, fmt.Errorf("ping failed to respond for redis server at %v", redisServer)
	}
	exists, err = r.client.SIsMember(MASTER_QUEUE_KEY, queueName).Result()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to identify queue '%s' on %v", redisServer)
	}
	if !exists && !create {
		return nil, fmt.Errorf("queue '%s' not found on %v", queueName, redisServer)
	}
	if exists {
		return r, nil
	}
	err = r.instantiate()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create named queue '%s' on %v", redisServer)
	}
	return r, nil
}

func (r *RedisQueue) Enqueue(payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal payload")
	}
	p := &QueuePayload{
		CreatedAt: time.Now(),
		JSON:      bytes,
	}
	data, err := p.Marshal()
	if err != nil {
		return errors.Wrapf(err, "failed to marshal queue entry")
	}
	lpush := r.client.LPush(r.queueName(), data)
	return lpush.Err()
}

func (r *RedisQueue) Dequeue(payload any, blocking bool) (bool, error) {
	var data string
	var err error
	for {
		data, err = r.client.LPop(r.queueName()).Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				if !blocking {
					return false, nil
				}
				time.Sleep(time.Duration(BLOCKING_DELAY_MSECS) * time.Millisecond)
				continue
			}
			return false, errors.Wrapf(err, "failed to pop queue entry")
		}
		break
	}
	entry := QueuePayload{}
	err = entry.Unmarshal(data)
	if err != nil {
		return false, errors.Wrapf(err, "failed to unmarshal queue entry")
	}
	err = json.Unmarshal(entry.JSON, payload)
	if err != nil {
		return false, errors.Wrapf(err, "failed to unmarshal payload")
	}
	return true, nil
}

func (r *RedisQueue) Close() error {
	err := r.client.Close()
	if err != nil {
		return errors.Wrapf(err, "failed to close redis client for queue %s", r.name)
	}
	return nil
}

func (r *RedisQueue) QueueLength() (int64, error) {
	len, err := r.client.LLen(r.queueName()).Result()
	if err != nil {
		return 0, errors.Wrapf(err, "failed to get length of queue %v", r.name)
	}
	return len, nil
}

func (r *RedisQueue) DestroyAndCleanup() error {
	err := r.client.Del(r.queueName()).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to cleanup existing items from queue %v", r.name)
	}
	return r.destroy()
}

func (r *RedisQueue) PingRedis() (bool, error) {
	statusCmd := r.client.Ping()
	res, err := statusCmd.Result()
	return res != "", err
}

func (r *RedisQueue) instantiate() error {
	err := r.client.SAdd(MASTER_QUEUE_KEY, r.name).Err()
	return err
}

func (r *RedisQueue) destroy() error {
	err := r.client.SRem(MASTER_QUEUE_KEY, r.name).Err()
	if err != nil {
		return errors.Wrapf(err, "failed to destroy queue %v", r.name)
	}
	return nil
}

func (r *RedisQueue) queueName() string {
	return MASTER_QUEUE_KEY + "::" + r.name
}
