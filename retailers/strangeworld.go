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
	SW_URL_PREFIX = "https://www.strangeworldrecords.com.au"
	SW_SEARCH_URL = "https://www.strangeworldrecords.com.au/search?q=%s+vinyl"
)

type StrangeWorldRecords struct{}

func (a *StrangeWorldRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(SW_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *StrangeWorldRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "\"grid__item one-fifth\"") >= 0 {

			title := toks[idx+1]
			if strings.Index(title, "- ") < 0 {
				continue
			}
			title = title[strings.Index(title, "title=\"")+7:]
			title = title[0:strings.Index(title, "\"")]
			url := toks[idx+1]
			url = url[strings.Index(url, "href=\"")+6:]
			url = SW_URL_PREFIX + url[0:strings.Index(url, "\"")]
			name := ""
			if strings.Index(strings.ToLower(title), "- ") > 0 {
				name = title[0:strings.Index(title, "- ")]
				name = strings.TrimSpace(name)
				title = title[strings.Index(title, "- ")+2:]
			}
			title = strings.TrimSpace(title)

			price := ""
			imageIdx := 2
			if strings.Index(toks[idx+2], "\"badge badge--sold-out\"") > 0 {
				price = SOLD_OUT
				imageIdx = 6
			} else {
				price = toks[idx+11]
				price = strings.TrimSpace(price[0:strings.Index(price, "</span")])
			}
			image := toks[idx+imageIdx]
			image = image[strings.Index(image, "src=\"")+5:]
			image = "https:" + image[0:strings.Index(image, "\"")]

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
