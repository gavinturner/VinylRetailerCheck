package retailers

import (
	"fmt"
)

type RetailerID int

const (
	// enum corresponding to retailer ids in the database
	ArtistFirst_Retailer RetailerID = 1
	PoisonCity_Retailer             = 2
)

type SKU struct {
	Name   string `json:"name"`
	Artist string `json:"artist"`
	Url    string `db:"item_url" json:"itemUrl"`
	Image  string `db:"image_url" json:"imageUrl"`
	Price  string `db:"price" json:"price"`
}

type VinylRetailer interface {
	GetArtistQueryURL(artist string) string
	ScrapeArtistReleases(artist string) ([]SKU, error)
}

func VinylRetailerFactory(retailerId RetailerID) (VinylRetailer, error) {
	switch retailerId {
	case ArtistFirst_Retailer:
		return &ArtistFirst{}, nil
	case PoisonCity_Retailer:
		return &PoisonCity{}, nil
	}
	return nil, fmt.Errorf("there is no retailer with id %v", retailerId)
}
