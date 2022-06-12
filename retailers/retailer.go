package retailers

import (
	"fmt"
)

type RetailerID int

const (
	SOLD_OUT = "sold out"
)

const (
	// enum corresponding to retailer ids in the database
	ArtistFirst_Retailer      RetailerID = 1
	PoisonCity_Retailer                  = 2
	ResistRecords_Retailer               = 3
	DamagedRecords_Retailer              = 4
	RepressedRecords_Retailer            = 5
	Utopia_Retailer                      = 6
	BeatdiscRecords_Retailer             = 7
	MusicFarmers_Retailer                = 8
)

type SKU struct {
	Name     string `db:"name" json:"name"`
	Artist   string `db:"artist" json:"artist"`
	Url      string `db:"item_url" json:"itemUrl"`
	Image    string `db:"image_url" json:"imageUrl"`
	Price    string `db:"price" json:"price"`
	Retailer string `db:"retailer" json:"retailer"`
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
	case ResistRecords_Retailer:
		return &ResistRecords{}, nil
	case DamagedRecords_Retailer:
		return &DamagedRecords{}, nil
	case RepressedRecords_Retailer:
		return &RepressedRecords{}, nil
	case Utopia_Retailer:
		return &Utopia{}, nil
	case BeatdiscRecords_Retailer:
		return &BeatDiscRecords{}, nil
	case MusicFarmers_Retailer:
		return &MusicFarmers{}, nil
	}

	return nil, fmt.Errorf("there is no retailer with id %v", retailerId)
}
