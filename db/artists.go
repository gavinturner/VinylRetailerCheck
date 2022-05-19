package db

import (
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"time"
)

type Artist struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	ImageUrl   string    `db:"image_url" json:"imageUrl"`
	WebsiteUrl string    `db:"website_url" json:"websiteUrl"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func (v *VinylDB) GetAllArtists(tx *postgres.Tx) ([]Artist, error) {
	querier := v.Q(tx)
	artists := []Artist{}
	err := querier.Select(&artists, `
		SELECT id, name, image_url, website_url, created_at, updated_at 
		FROM artists
	`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve artists")
	}
	return artists, nil
}
