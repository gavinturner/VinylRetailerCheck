package retailers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// https://repressedrecords.com/search?q=clowns+vinyl&options%5Bprefix%5D=last

const (
	RER_URL_PREFIX = "https://repressedrecords.com"
	RER_SEARCH_URL = "https://repressedrecords.com/search?q=%s+vinyl&options[prefix]=last"
)

type RepressedRecords struct{}

func (a *RepressedRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(RER_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *RepressedRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "class=\"h4 grid-view-item__title product-card__title\"") >= 0 {

			title := toks[idx+1]
			title = title[0:strings.Index(title, "</div")]
			name := ""
			if strings.Index(strings.ToLower(title), " - ") > 0 {
				name = title[0:strings.Index(title, " - ")]
				title = title[strings.Index(title, " - ")+3:]
			}
			price := toks[idx+17]
			price = strings.TrimSpace(price[0:strings.Index(price, "</span")])
			image := toks[idx-2]
			image = image[strings.Index(image, "src=\"")+5:]
			image = "http:" + image[0:strings.Index(image, "\"")]
			url := toks[idx-15]
			url = url[strings.Index(url, "href=")+6:]
			url = RER_URL_PREFIX + url[0:strings.Index(url, "\"")]
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
