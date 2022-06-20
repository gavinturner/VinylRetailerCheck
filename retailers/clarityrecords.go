package retailers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// https://clarityrecords.net/search.php?search_query=clowns+vinyl&section=product

const (
	CR_URL_PREFIX = "https://clarityrecords.net"
	CR_SEARCH_URL = "https://clarityrecords.net/search.php?search_query=%s+vinyl&section=product"
)

type ClarityRecords struct{}

func (a *ClarityRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(CR_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *ClarityRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "data-name=\"") >= 0 {
			title := t[strings.Index(t, "data-name=\"")+11:]
			title = title[0:strings.Index(title, "\"")]
			name := ""
			if strings.Index(strings.ToLower(title), strings.ToLower(artist+" - ")) >= 0 {
				name = title[strings.Index(strings.ToLower(title), strings.ToLower(artist+" - ")):strings.Index(title, " - ")]
				title = title[strings.Index(title, " - ")+3:]
			}

			price := t[strings.Index(t, "data-product-price=\"")+20:]
			price = "$" + price[0:strings.Index(price, "\"")]
			url := toks[idx+2]
			url = url[strings.Index(url, "href=\"")+6:]
			url = url[0:strings.Index(url, "\"")]
			image := toks[idx+4]
			image = image[strings.Index(image, "data-src=\"")+10:]
			image = image[0:strings.Index(image, "\"")]

			soldOut := toks[idx+12]
			if strings.Index(soldOut, "Sold out") >= 0 {
				price = SOLD_OUT
			}

			sku := SKU{
				Url:    url,
				Artist: strings.ToLower(strings.TrimSpace(name)),
				Name:   strings.TrimSpace(title),
				Price:  strings.TrimSpace(price),
				Image:  image,
			}

			// artist name can contain the searched for artist if it's a split etc.
			if sku.Artist != strings.ToLower(artist) {
				continue
			}
			findings = append(findings, sku)
		}
	}
	return findings, nil
}
