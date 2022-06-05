package retailers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// https://damagedmusic.com.au/?s=clowns+vinyl&post_type=product

const (
	DR_SEARCH_URL = "https://damagedmusic.com.au/?s=%s+vinyl&post_type=product"
)

type DamagedRecords struct{}

func (a *DamagedRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(DR_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *DamagedRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	// note if damaged has only one result then it goes directly to the product page..

	findings = []SKU{}
	query := a.GetArtistQueryURL(artist)
	resp, err := http.Get(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve search query %s", query)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract body of search results %s", query)
	}

	toks := strings.Split(string(body), ">")
	for idx, t := range toks {
		if strings.Index(t, "class=\"product-wrap\"") >= 0 {
			title := toks[idx+1]
			title = title[strings.Index(title, "aria-label=")+12:]
			title = title[0:strings.Index(title, "\"")]
			name := title[0:strings.Index(title, " - ")]
			title = title[strings.Index(title, " - ")+3:]
			image := toks[idx+2]
			image = image[strings.Index(image, "src=\"")+5:]
			image = image[0:strings.Index(image, "\"")]
			image = "<img width=\"150px\" height=\"150px\" src=\"" + image + "\">"
			url := toks[idx+5]
			url = url[strings.Index(url, "href=")+6:]
			url = url[0:strings.Index(url, "\"")]

			price := toks[idx+15]
			if end := strings.Index(price, "</bdi"); end > 0 {
				price = "$" + price[0:end]
			}
			sku := SKU{
				Url:    url,
				Artist: strings.ToLower(strings.TrimSpace(name)),
				Name:   strings.TrimSpace(title),
				Price:  strings.TrimSpace(price),
				Image:  image,
			}

			// artist name can contain the searched for artist if it's a split etc.
			if strings.Index(sku.Artist, strings.ToLower(artist)) < 0 {
				continue
			}
			findings = append(findings, sku)
		}
	}
	return findings, nil
}
