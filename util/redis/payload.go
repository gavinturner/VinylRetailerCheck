package redis

import (
	"encoding/json"
	"github.com/pkg/errors"
	"time"
)

type QueuePayload struct {
	JSON      []byte
	CreatedAt time.Time
}

func (p *QueuePayload) Marshal() (string, error) {
	bytes, err := json.Marshal(*p)
	if err != nil {
		return "", errors.Wrapf(err, "failed to marshal json")
	}
	return string(bytes), nil
}

func (p *QueuePayload) Unmarshal(data string) error {
	err := json.Unmarshal([]byte(data), p)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal json")
	}
	return nil
}
