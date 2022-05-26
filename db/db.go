package db

import (
	"fmt"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"time"
)

const (
	DB_WAIT_DELAY_MSECS = 500
)

// generate an interface and mock from our persistor implementation
//go:generate ifacemaker --sort=true -f "*.go" -s VinylDB -i VinylDS -p db -o vinylds.go
//go:generate goimports -w vinylds.go
//go:generate sed -i -e  /null.\"gopkg\.in\/guregu\/null\.v3\"/d vinylds.go
//go:generate rm -f vinylds_mock.go
//go:generate moq -out vinylds_mock.go . VinylDS

type VinylDB struct {
	db *postgres.DB
}

func NewDB(db *postgres.DB) VinylDB {
	return VinylDB{
		db: db,
	}
}

func (v *VinylDB) Q(tx *postgres.Tx) postgres.Querier {
	var querier postgres.Querier = v.db
	if tx != nil {
		querier = tx
	}
	return querier
}

func (v *VinylDB) VerifySchema() error {
	var versions []int64
	err := v.db.Select(&versions, `SELECT version from schema_migrations`)
	if err != nil {
		return errors.Wrapf(err, "Failed to retrieve current scheme version")
	}
	if len(versions) == 0 {
		return fmt.Errorf("Schema migration has not been run")
	}
	return nil
}

func (v *VinylDB) WaitForDbUp(timeoutSecs int64) error {
	var delay int64
	var err error
	for {
		err = v.VerifySchema()
		if err == nil {
			break
		}
		time.Sleep(DB_WAIT_DELAY_MSECS * time.Millisecond)
		delay += DB_WAIT_DELAY_MSECS
		if delay > timeoutSecs*1000 {
			return fmt.Errorf("Timeout waiting for db to come up")
		}
	}
	log.Debugf("Database looks ok..\n")
	return nil
}
