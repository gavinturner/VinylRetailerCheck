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
type WatchedArtist struct {
	ArtistID   int64  `db:"artist_id" json:"artistId"`
	ArtistName string `db:"artist_name" json:"artistName"`
	UserID     int64  `db:"user_id" json:"userId"`
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

func (v *VinylDB) GetWatchedArtists(tx *postgres.Tx) (map[int64][]WatchedArtist, error) {
	querier := v.Q(tx)
	watched := []WatchedArtist{}
	err := querier.Select(&watched, `
		SELECT fa.user_id, fa.artist_id, a.name as artist_name
		FROM users_following_artists fa
		JOIN artists a ON fa.artist_id = a.id
	`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve watched artist lists")
	}
	cached := map[int64][]WatchedArtist{}
	for _, watchItem := range watched {
		if _, ok := cached[watchItem.UserID]; !ok {
			cached[watchItem.UserID] = []WatchedArtist{}
		}
		cached[watchItem.UserID] = append(cached[watchItem.UserID], watchItem)
	}
	return cached, nil
}
