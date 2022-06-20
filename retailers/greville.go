package retailers

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// https://grevillerecords.com.au/search?options%5Bprefix%5D=last&page=1&q=pixies+vinyl
// dont forget to go though the pages of results

const (
	GR_URL_PREFIX = "https://grevillerecords.com.au"
	GR_SEARCH_URL = "https://grevillerecords.com.au/search?options[prefix]=last&page=%v&q=%s+vinyl"
)

type GrevilleRecords struct{}

func (a *GrevilleRecords) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(GR_SEARCH_URL, "1", url.QueryEscape(artist))
	return query
}

func (a *GrevilleRecords) GetArtistQueryForPageURL(artist string, page int) string {
	query := fmt.Sprintf(GR_SEARCH_URL, page, url.QueryEscape(artist))
	return query
}

type GrevilleInfo struct {
	Id                   int64       `json:"id"`
	Title                string      `json:"title"`
	Handle               string      `json:"handle"`
	Description          string      `json:"description"`
	PublishedAt          time.Time   `json:"published_at"`
	CreatedAt            time.Time   `json:"created_at"`
	Vendor               string      `json:"vendor"`
	Type                 string      `json:"type"`
	Tags                 []string    `json:"tags"`
	Price                int         `json:"price"`
	PriceMin             int         `json:"price_min"`
	PriceMax             int         `json:"price_max"`
	Available            bool        `json:"available"`
	PriceVaries          bool        `json:"price_varies"`
	CompareAtPrice       interface{} `json:"compare_at_price"`
	CompareAtPriceMin    int         `json:"compare_at_price_min"`
	CompareAtPriceMax    int         `json:"compare_at_price_max"`
	CompareAtPriceVaries bool        `json:"compare_at_price_varies"`
	Variants             []struct {
		Id                     int64         `json:"id"`
		Title                  string        `json:"title"`
		Option1                string        `json:"option1"`
		Option2                interface{}   `json:"option2"`
		Option3                interface{}   `json:"option3"`
		Sku                    string        `json:"sku"`
		RequiresShipping       bool          `json:"requires_shipping"`
		Taxable                bool          `json:"taxable"`
		FeaturedImage          interface{}   `json:"featured_image"`
		Available              bool          `json:"available"`
		Name                   string        `json:"name"`
		PublicTitle            interface{}   `json:"public_title"`
		Options                []string      `json:"options"`
		Price                  int           `json:"price"`
		Weight                 int           `json:"weight"`
		CompareAtPrice         interface{}   `json:"compare_at_price"`
		InventoryManagement    string        `json:"inventory_management"`
		Barcode                string        `json:"barcode"`
		RequiresSellingPlan    bool          `json:"requires_selling_plan"`
		SellingPlanAllocations []interface{} `json:"selling_plan_allocations"`
	} `json:"variants"`
	Images        []string `json:"images"`
	FeaturedImage string   `json:"featured_image"`
	Options       []string `json:"options"`
	Media         []struct {
		Alt          interface{} `json:"alt"`
		Id           int64       `json:"id"`
		Position     int         `json:"position"`
		PreviewImage struct {
			AspectRatio float32 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			Width       int     `json:"width"`
			Src         string  `json:"src"`
		} `json:"preview_image"`
		AspectRatio float32 `json:"aspect_ratio"`
		Height      int     `json:"height"`
		MediaType   string  `json:"media_type"`
		Src         string  `json:"src"`
		Width       int     `json:"width"`
	} `json:"media"`
	RequiresSellingPlan bool          `json:"requires_selling_plan"`
	SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
	Content             string        `json:"content"`
}

func (a *GrevilleRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	page := 1
	findings = []SKU{}
	for {
		found := false
		query := a.GetArtistQueryForPageURL(artist, page)
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
			if strings.Index(t, "\"product-card product-card--list\"") >= 0 {

				url := toks[idx+1]
				url = url[strings.Index(url, "href=\"")+6:]
				url = url[0:strings.Index(url, "\"")]
				url = GR_URL_PREFIX + url

				// read the detailed page to get the json data
				resp, err := http.Get(url)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to retrieve json body detais %s", url)
				}
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to extract body of title details %s", query)
				}

				var price, image, title, name string
				subToks := strings.Split(string(body), ">")
				for subIdx, tok := range subToks {
					// search for the json data
					START_TOKEN := "\"ProductJson-product-template\""
					END_TOKEN := "</script"

					if strings.Index(tok, START_TOKEN) >= 0 {
						jd := subToks[subIdx+1]
						jd = strings.TrimSpace(jd[0:strings.Index(jd, END_TOKEN)])
						data := GrevilleInfo{}
						err := json.Unmarshal([]byte(jd), &data)
						if err != nil {
							fmt.Printf("[FAILED: %s]..", err.Error())
							return nil, errors.Wrapf(err, "failed to parse json data")
						}
						if !data.Available {
							price = SOLD_OUT
						} else {
							price = fmt.Sprintf("$%.2f", float32(data.Price)/100.0)
						}
						image = "https:" + data.FeaturedImage
						title = data.Title
						if strings.Index(title, "- ") > 0 {
							name = strings.TrimSpace(title[0:strings.Index(title, "- ")])
							title = strings.TrimSpace(title[strings.Index(title, "- ")+2:])
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
				found = true

				// artist name can contain the searched for artist if it's a split etc.
				if sku.Artist != strings.ToLower(artist) {
					continue
				}
				findings = append(findings, sku)
			}
		}
		if !found {
			break
		}
		page += 1
	}
	return findings, nil
}
