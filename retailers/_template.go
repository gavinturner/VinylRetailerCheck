package retailers

import (
	"fmt"
	"net/url"
)

// https://repressedrecords.com/search?q=clowns+vinyl&options%5Bprefix%5D=last

const (
	TEMP_URL_PREFIX = "https://repressedrecords.com"
	TEMP_SEARCH_URL = "https://repressedrecords.com/search?q=%s+vinyl&options%5Bprefix%5D=last"
)

type Template struct{}

func (a *Template) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(TEMP_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *Template) ScrapeArtistReleases(artist string) (findings []SKU, err error) {
	return nil, nil
}
