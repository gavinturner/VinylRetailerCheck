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

// https://damagedmusic.com.au/?s=clowns+vinyl&post_type=product

const (
	MF_URL_PREFIX = "https://musicfarmers.com"
	MF_SEARCH_URL = "https://musicfarmers.com/search?q=%s"
)

type MusicFarmers struct{}
type MusicFarmersInfo struct {
	Id                   int64         `json:"id"`
	Title                string        `json:"title"`
	Handle               string        `json:"handle"`
	Description          string        `json:"description"`
	PublishedAt          time.Time     `json:"published_at"`
	CreatedAt            time.Time     `json:"created_at"`
	Vendor               string        `json:"vendor"`
	Type                 string        `json:"type"`
	Tags                 []interface{} `json:"tags"`
	Price                int           `json:"price"`
	PriceMin             int           `json:"price_min"`
	PriceMax             int           `json:"price_max"`
	Available            bool          `json:"available"`
	PriceVaries          bool          `json:"price_varies"`
	CompareAtPrice       interface{}   `json:"compare_at_price"`
	CompareAtPriceMin    int           `json:"compare_at_price_min"`
	CompareAtPriceMax    int           `json:"compare_at_price_max"`
	CompareAtPriceVaries bool          `json:"compare_at_price_varies"`
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
		InventoryQuantity      int           `json:"inventory_quantity"`
		InventoryManagement    string        `json:"inventory_management"`
		InventoryPolicy        string        `json:"inventory_policy"`
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

func (a *MusicFarmers) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(MF_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *MusicFarmers) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	// TODO: if damaged has only one result then it goes directly to the product page..

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
		if strings.Index(t, "grid__item one-fifth") >= 0 {
			title := toks[idx+1]
			title = title[strings.Index(title, "title=\"")+7:]
			title = title[0:strings.Index(title, "\"")]
			url := toks[idx+1]
			url = url[strings.Index(url, "href=\"")+6:]
			url = url[0:strings.Index(url, "\"")]
			url = MF_URL_PREFIX + url
			image := toks[idx+2]
			image = image[strings.Index(image, "src=\"")+5:]
			image = "https:" + image[0:strings.Index(image, "\"")]

			sku := SKU{
				Url:   url,
				Name:  strings.TrimSpace(title),
				Image: image,
			}

			// now load the url and get some more details..
			resp, err := http.Get(url)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to retrieve sub query %s", url)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to extract body of sub query results %s", url)
			}

			subToks := strings.Split(string(body), ">")
			data := MusicFarmersInfo{}
			for _, t := range subToks {
				if strings.Index(t, "Shopify.OptionSelectors('productSelect',") >= 0 {
					jd := t[strings.Index(t, "product: ")+9:]
					jd = jd[0 : strings.Index(jd, "onVariantSelected: selectCallback")-8]
					err := json.Unmarshal([]byte(jd), &data)
					if err != nil {
						return nil, errors.Wrapf(err, "failed to parse body of sub query results %s", url)
					}
					break
				}
			}

			sku.Name = data.Title
			sku.Artist = strings.ToLower(data.Vendor)
			sku.Price = fmt.Sprintf("$0.2f", float32(data.Price)/100.0)
			if !data.Available {
				sku.Price = SOLD_OUT
			}

			// artist name can contain the searched for artist if it's a split etc.
			if strings.ToLower(sku.Artist) != strings.ToLower(artist) {
				continue
			}
			findings = append(findings, sku)
		}
	}
	return findings, nil
}
