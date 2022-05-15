package retailers

import (
	"github.com/gavinturner/VinylRetailChecker/db"
	"github.com/gavinturner/VinylRetailChecker/files"
)

type VinylRetailer interface {
	GetArtistQueryURL(artist string) string
	ScrapeArtistReleases(artist string, prices *files.PricesStore) (findings []db.SKU, found bool)
}
