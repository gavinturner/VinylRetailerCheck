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

const (
	DV_URL_PREFIX = "https://www.dutchvinyl.com.au"
	DV_SEARCH_URL = "https://www.dutchvinyl.com.au/a/search?type=product&q=%s+vinyl"
)

type DutchVinyl struct{}

func (a *DutchVinyl) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(DV_SEARCH_URL, url.QueryEscape(artist))
	return query
}

type DutchVinylInfo struct {
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
		Alt          string `json:"alt"`
		Id           int64  `json:"id"`
		Position     int    `json:"position"`
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

func (a *DutchVinyl) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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

	for _, t := range toks {
		if strings.Index(t, "class=\"productitem--link\"") >= 0 {
			sku := SKU{}
			subUrl := t[strings.Index(t, "href=\"")+6:]
			subUrl = subUrl[0:strings.Index(subUrl, "\"")]
			subUrl = DV_URL_PREFIX + subUrl

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
				START_TOKEN := "\"product\": "
				END_TOKEN := "\n"

				if strings.Index(tok, START_TOKEN) >= 0 {
					jd := tok[strings.Index(tok, START_TOKEN)+len(START_TOKEN):]
					jd = jd[0:strings.Index(jd, END_TOKEN)]
					data := DutchVinylInfo{}
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
			}
			// "Pixies – Debaser (Demo) (VG+/VG+)"
			if strings.Index(strings.ToLower(sku.Name), strings.ToLower(artist)+" –") < 0 &&
				strings.Index(strings.ToLower(sku.Name), "– "+strings.ToLower(artist)) < 0 {
				continue
			}
			sku.Url = subUrl

			if strings.Index(strings.ToLower(sku.Name), strings.ToLower(artist+" -")) == 0 {
				sku.Name = sku.Name[len(artist)+2:]
				sku.Name = strings.TrimSpace(sku.Name)
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
