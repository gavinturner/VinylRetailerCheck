package db

import (
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/pkg/errors"
	"time"
)

type Release struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	ArtistID  int64     `db:"artist_id" json:"artistID"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (v *VinylDB) UpsertRelease(tx *postgres.Tx, artistId int64, title string) (id int64, err error) {
	querier := v.Q(tx)
	err = querier.Get(&id, querier.Rebind(`
		INSERT INTO releases (title, artist_id) VALUES (?, ?)
		ON CONFLICT (artist_id, title) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`), title, artistId)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to upsert release")
	}
	return id, err
}
