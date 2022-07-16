package retailers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	AF_URL_PREFIX = "https://artistfirst.com.au"
	AF_SEARCH_URL = "https://artistfirst.com.au/search?q=%s+vinyl"
)

type ArtistFirst struct{}

func (a *ArtistFirst) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(AF_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *ArtistFirst) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "product-list-item-title") >= 0 {
			sku := SKU{
				Url:    strings.Replace(toks[idx+1], "<a href=\"", AF_URL_PREFIX, -1),
				Artist: strings.ToLower(strings.TrimSpace(strings.TrimSuffix(toks[idx-1], "</p"))),
				Image:  strings.TrimSpace(toks[idx-10]),
				Name:   strings.TrimSpace(strings.TrimSuffix(toks[idx+2], "</a")),
				Price:  toks[idx+5], // sold out
			}
			if sku.Artist == "" {
				sku.Artist = strings.ToLower(strings.TrimSpace(strings.TrimSuffix(toks[idx-2], "</a")))
			}

			sku.Artist = strings.TrimSpace(strings.Replace(sku.Artist, "pre-order", "", -1))

			sku.Image = sku.Image[strings.Index(sku.Image, "src=\""):]
			sku.Image = strings.Replace(sku.Image, "src=\"", fmt.Sprintf("https:"), -1)
			if strings.Index(strings.ToLower(sku.Price), SOLD_OUT) < 0 {
				// ok we didnt find the sold out price
				sku.Price = toks[idx+6] // price?

				if strings.Index(strings.ToLower(sku.Price), "class=\"money") >= 0 {
					// ok we found the 'on special' case - one more..
					sku.Price = toks[idx+7] // price (when on sale)
					sku.Image = strings.TrimSpace(toks[idx-10])
				} else {
					sku.Image = strings.TrimSpace(toks[idx-8])
				}
				sku.Image = sku.Image[strings.Index(sku.Image, "src=\""):]
				sku.Image = strings.Replace(sku.Image, "src=\"", fmt.Sprintf("https:"), -1)
				sku.Image = sku.Image[0 : strings.Index(sku.Image, "\"")-1]
				sku.Price = strings.TrimSuffix(sku.Price, "</span")
			} else {
				sku.Price = SOLD_OUT
			}
			sku.Image = sku.Image[0:strings.Index(sku.Image, "\"")]
			sku.Price = strings.TrimSpace(sku.Price)

			if sku.Artist != artist {
				continue
			}
			findings = append(findings, sku)
		}
	}
	return findings, nil
}
