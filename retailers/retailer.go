package retailers

import (
	"github.com/gavinturner/VinylRetailChecker/db"
)

type VinylRetailer interface {
	GetArtistQueryURL(artist string) string
	ScrapeArtistReleases(artist string) ([]db.SKU, error)
}
