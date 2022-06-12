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

const (
	OFFW_URL_PREFIX = "https://offwhiterecords.com/"
	OFFW_SEARCH_URL = "https://offwhiterecords.com/store/search_results.cfm?search_term=%s&search_sub=SEARCH"
)

type OffWhiteRecords struct{}

func (a *OffWhiteRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(OFFW_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *OffWhiteRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	findings = []SKU{}
	query := a.GetArtistQueryURL(artist)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", query, nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve search query %s", query)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract body of search results %s", query)
	}
	toks := strings.Split(string(body), ">")

	for idx, t := range toks {
		if strings.Index(t, "<tr class=\"altrow\"") >= 0 {
			sku := SKU{}
			subUrl := strings.TrimPrefix(strings.TrimSpace(toks[idx+7]), "<a href=\"")
			subUrl = PC_URL_PREFIX + strings.TrimSuffix(subUrl, "\"")

			// now read THAT url...
			query := fmt.Sprintf(subUrl)
			resp, err := http.Get(query)
			if err != nil {
				fmt.Printf("[FAILED: %s]..", err.Error())
				return nil, errors.Wrapf(err, "failed to open sub url %s", subUrl)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("[FAILED: %s]..", err.Error())
				return nil, errors.Wrapf(err, "failed to open body of sub url %s", subUrl)
			}

			subToks := strings.Split(string(body), ">")
			for _, tok := range subToks {

				// search for the json data
				START_TOKEN := "new Shopify.OptionSelectors(\"product-select\", { product:"
				END_TOKEN := ", onVariantSelected"

				if strings.Index(tok, START_TOKEN) >= 0 {
					jd := tok[strings.Index(tok, START_TOKEN)+len(START_TOKEN):]
					jd = jd[0:strings.Index(jd, END_TOKEN)]
					data := PoisonCityInfo{}
					err := json.Unmarshal([]byte(jd), &data)
					if err != nil {
						fmt.Printf("[FAILED: %s]..", err.Error())
						return nil, errors.Wrapf(err, "failed to parse json data")
					}

					sku.Name = data.Title
					sku.Artist = artist
					sku.Image = "https:" + data.FeaturedImage
					sku.Price = fmt.Sprintf("$%.2f", float32(data.Price)/100.0)
					if !data.Available {
						sku.Price = SOLD_OUT
					}
					break
				}

				/*
					if strings.Index(tok, "itemprop=\"image\"") > 0 {
						sku.Image = "https:" + strings.TrimPrefix(strings.TrimSpace(tok), "<meta itemprop=\"image\" content=\"")
						sku.Image = strings.TrimSuffix(sku.Image, "\" /")
					}
					if strings.Index(tok, "itemprop=\"name\"") >= 0 {
						sku.Name = strings.TrimSpace(subToks[idx+1])
						sku.Name = strings.TrimSuffix(sku.Name, "</h1")
					}
					if strings.Index(tok, "itemprop=\"price\"") >= 0 {
						sku.Price = strings.TrimSpace(subToks[idx+1])
						sku.Price = strings.TrimSuffix(sku.Price, "</span")
					}
					if strings.Index(tok, "\"Sold Out\"") >= 0 {
						sku.Price = SOLD_OUT
					}*/
			}

			if !strings.HasPrefix(strings.ToLower(sku.Name), strings.ToLower(artist)) {
				continue
			}
			sku.Url = subUrl

			if strings.Index(strings.ToLower(sku.Name), strings.ToLower(artist+" ")) == 0 {
				sku.Name = sku.Name[len(artist)+1:]
			}
			if (sku.Name[0] == '\'' && sku.Name[len(sku.Name)-1] == '\'') ||
				(sku.Name[0] == '"' && sku.Name[len(sku.Name)-1] == '"') {
				sku.Name = sku.Name[1 : len(sku.Name)-1]
			}
			// store new price
			findings = append(findings, sku)
		}
	}
	return findings, nil
}
