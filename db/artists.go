package db

import (
	"github.com/gavinturner/vinylretailers/util/postgres"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"time"
)

type Artist struct {
	ID         int64     `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	ImageUrl   string    `db:"image_url" json:"imageUrl"`
	WebsiteUrl string    `db:"website_url" json:"websiteUrl"`
	Variants   []string  `db:"variants" json:"variants"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}
type WatchedArtist struct {
	ArtistID       int64    `db:"artist_id" json:"artistId"`
	ArtistName     string   `db:"artist_name" json:"artistName"`
	ArtistVariants []string `db:"artist_variants" json:"artistVariants"`
	UserID         int64    `db:"user_id" json:"userId"`
}

func (v *VinylDB) GetAllArtists(tx *postgres.Tx) ([]Artist, error) {
	querier := v.Q(tx)
	artists := []Artist{}
	err := querier.Select(&artists, `
		SELECT id, name, variants, image_url, website_url, created_at, updated_at 
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
	rows, err := querier.Query(`
		SELECT fa.user_id, fa.artist_id, a.name as artist_name, a.variants as artist_variants
		FROM users_following_artists fa
		JOIN artists a ON fa.artist_id = a.id
	`)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve watched artist lists")
	}
	defer rows.Close()
	for rows.Next() {
		var a WatchedArtist
		err = rows.Scan(&a.UserID, &a.ArtistID, &a.ArtistName, pq.Array(a.ArtistVariants))
		if err != nil {
			return nil, errors.Wrapf(err, "failed to scan watched artist row")
		}
		watched = append(watched, a)
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
