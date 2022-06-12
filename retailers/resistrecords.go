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
	RR_URL_PREFIX = "https://shop.resistrecords.com"
	RR_SEARCH_URL = "https://shop.resistrecords.com/search?q=%s+vinyl"
)

type ResistRecords struct{}

func (a *ResistRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(RR_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *ResistRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "class=\"title\"") >= 0 {
			url := t[strings.Index(t, "href=\"")+6:]
			url = url[0:strings.Index(url, "\"")]
			description := toks[idx+1]
			name := "Various"
			title := ""
			if strings.Index(description, "\"") > 0 {
				name = description[0:strings.Index(description, "\"")]
				title = description[len(name):]
			} else {
				title = description
			}
			title = title[0:strings.Index(title, "</a")]

			image := toks[idx-8]
			image = image[strings.Index(image, "src=\"")+5:]
			image = "https:" + image[0:strings.Index(image, "\"")]
			image = strings.Replace(image, "_120x", "_240x", 1)

			price := toks[idx+4]
			if strings.Index(price, "sold-out") >= 0 {
				price = SOLD_OUT
			}
			if strings.Index(price, "sale") >= 0 {
				price = toks[idx+7]
				if strings.Index(price, "From") >= 0 {
					price = toks[idx+10]
				}
			}
			if strings.Index(price, "</span") > 0 {
				price = price[0:strings.Index(price, "</span")]
			}
			sku := SKU{
				Url:    RR_URL_PREFIX + url,
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
