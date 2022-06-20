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
	Retailer_ArtistFirst         RetailerID = 1
	Retailer_PoisonCity                     = 2
	Retailer_ResistRecords                  = 3
	Retailer_DamagedRecords                 = 4
	Retailer_RepressedRecords               = 5
	Retailer_Utopia                         = 6
	Retailer_BeatdiscRecords                = 7
	Retailer_MusicFarmers                   = 8
	Retailer_DutchVinyl                     = 9
	Retailer_ClarityRecords                 = 10
	Retailer_OhJeanRecords                  = 11
	Retailer_StrengeWorldRecords            = 12
	Retailer_GrevilleRecords                = 13
)

type SKU struct {
	Name        string `db:"name" json:"name"`
	Artist      string `db:"artist" json:"artist"`
	Url         string `db:"item_url" json:"itemUrl"`
	Image       string `db:"image_url" json:"imageUrl"`
	Price       string `db:"price" json:"price"`
	Retailer    string `db:"retailer" json:"retailer"`
	RetailerUrl string `db:"retailer_url" json:"retailerUrl"`
}

type VinylRetailer interface {
	GetArtistQueryURL(artist string) string
	ScrapeArtistReleases(artist string) ([]SKU, error)
}

func VinylRetailerFactory(retailerId RetailerID) (VinylRetailer, error) {
	switch retailerId {
	case Retailer_ArtistFirst:
		return &ArtistFirst{}, nil
	case Retailer_PoisonCity:
		return &PoisonCity{}, nil
	case Retailer_ResistRecords:
		return &ResistRecords{}, nil
	case Retailer_DamagedRecords:
		return &DamagedRecords{}, nil
	case Retailer_RepressedRecords:
		return &RepressedRecords{}, nil
	case Retailer_Utopia:
		return &Utopia{}, nil
	case Retailer_BeatdiscRecords:
		return &BeatDiscRecords{}, nil
	case Retailer_MusicFarmers:
		return &MusicFarmers{}, nil
	case Retailer_DutchVinyl:
		return &DutchVinyl{}, nil
	case Retailer_ClarityRecords:
		return &ClarityRecords{}, nil
	case Retailer_OhJeanRecords:
		return &OhJeanRecords{}, nil
	case Retailer_StrengeWorldRecords:
		return &StrangeWorldRecords{}, nil
	case Retailer_GrevilleRecords:
		return &GrevilleRecords{}, nil
	}

	return nil, fmt.Errorf("there is no retailer with id %v", retailerId)
}
