package db

import (
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"time"
)

type Retailer struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Url       string    `db:"url" json:"url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (v *VinylDB) GetAllRetailers(tx *postgres.Tx) ([]Retailer, error) {
	querier := v.Q(tx)
	retailers := []Retailer{}
	err := querier.Select(&retailers, `
		SELECT id, name, url, created_at, updated_at 
		FROM retailers
	`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve retailers")
	}
	return retailers, nil
}
