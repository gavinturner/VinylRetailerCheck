package retailers

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// https://www.ohjeanrecords.com/apps/omega-search/?type=product&options[prefix]=last&q=%s

const (
	OJ_URL_PREFIX = "https://www.ohjeanrecords.com"
	OJ_SEARCH_URL = "https://www.ohjeanrecords.com/apps/omega-search/?type=product&options%%5Bprefix%%5D=last&q=%s"
)

type OhJeanRecords struct{}

type OhJaanInfo struct {
	Id       int64  `json:"id"`
	Gid      string `json:"gid"`
	Vendor   string `json:"vendor"`
	Type     string `json:"type"`
	Variants []struct {
		Id          int64       `json:"id"`
		Price       int         `json:"price"`
		Name        string      `json:"name"`
		PublicTitle interface{} `json:"public_title"`
		Sku         string      `json:"sku"`
	} `json:"variants"`
}

func (a *OhJeanRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(OJ_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *OhJeanRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "<a class=\"os-e os-name\"") >= 0 {
			url := t[strings.Index(t, "href=\"")+6:]
			url = url[0:strings.Index(url, "\"")]
			url = OJ_URL_PREFIX + url
			var name, price, image string
			title := toks[idx+1]
			if strings.Index(title, "<span class=\"omega__highlight\"") >= 0 {
				name = toks[idx+2]
				if strings.Index(name, "<span class=\"omega__highlight\"") >= 0 {
					name = toks[idx+3]
					image = toks[idx-4]
					price = toks[idx+7]
				} else {
					image = toks[idx-4]
					price = toks[idx+5]
				}
				name = strings.TrimSpace(name[0:strings.Index(name, "</span")])
				title = strings.TrimSpace(title[0:strings.Index(title, "<span class=\"omega__highlight\"")])
				if strings.Index(title, " -") > 0 {
					title = title[0:strings.Index(title, " -")]
				}

			} else {
				title = title[0:strings.Index(title, "</a")]
				image = toks[idx-4]
				price = toks[idx+3]
			}
			image = image[strings.Index(image, "src=\"")+5:]
			image = "https:" + image[0:strings.Index(image, "\"")]
			price = strings.TrimSpace(price[0:strings.Index(price, "</div")])
			if strings.Index(price, "sold-out") >= 0 {
				price = SOLD_OUT
			}

			// now read the subUrl to check it's the correct artist
			query := url
			resp, err := http.Get(query)
			if err != nil {
				fmt.Printf("[FAILED: %s]..", err.Error())
				return nil, errors.Wrapf(err, "failed to open sub url %s", url)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("[FAILED: %s]..", err.Error())
				return nil, errors.Wrapf(err, "failed to open body of sub url %s", url)
			}

			subToks := strings.Split(string(body), ">")
			for _, tok := range subToks {
				// search for the json data
				START_TOKEN := "var meta = {\"product\":"
				END_TOKEN := ",\"page\":{"

				if strings.Index(tok, START_TOKEN) >= 0 {
					jd := tok[strings.Index(tok, START_TOKEN)+len(START_TOKEN):]
					jd = jd[0:strings.Index(jd, END_TOKEN)]
					data := OhJaanInfo{}
					err := json.Unmarshal([]byte(jd), &data)
					if err != nil {
						fmt.Printf("[FAILED: %s]..", err.Error())
						return nil, errors.Wrapf(err, "failed to parse json data")
					}
					name = data.Vendor
					name = strings.TrimSpace(name)
					if len(name) > len(artist) {
						name = name[0:len(artist)]
					}
					if len(data.Variants) > 0 {
						title = strings.TrimSpace(data.Variants[0].Name)
					}
				}
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
