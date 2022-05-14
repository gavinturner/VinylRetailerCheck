package main

type SKU struct {
	Url    string
	Artist string
	Image  string
	Name   string
	Price  string
}

type VinylRetailer interface {
	GetArtistQueryURL(artist string) string
	ScrapeArtistReleases(artist string, prices *KnownPrices) (findings []SKU, found bool)
}
