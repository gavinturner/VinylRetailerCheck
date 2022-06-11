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
	PC_URL_PREFIX = "https://poisoncityestore.com"
	PC_SEARCH_URL = "https://poisoncityestore.com/search?q=%s+vinyl"
)

type PoisonCity struct{}

// search for: new Shopify.OptionSelectors("product-select",{ product: { ... }, onVariantSelected: selectCallback });
// we want the product json part
type PoisonCityInfo struct {
	Id                   int         `json:"id"`
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
		Id                     int           `json:"id"`
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

/*
{
  "id": 328995763,
  "title": "DESCENDENTS 'Everything Sucks' LP",
  "handle": "descendents-everything-sucks-lp",
  "description": "<span>BACK IN STOCK! On an undisclosed date in 1996, deemed by some as Science's Finest Hour, Milo hung up his lab coat and the Descendents began work on a new LP. Produced by Bill and Stephen at the band's own studio in Ft. Collins, CO (aptly dubbed The Blasting Room), \"Everything Sucks\" picks up EXACTLY where the Descendents left off. The tracks \"Everything Sucks,\" \"This Place\" and \"Coffee Mug\" are classic Descendents, raging along at the speed of sound, while \"I'm The One\", \"Sick-O-Me\" and \"She Loves Me\" are anthems of the \"girl-song\" milieu.</span>",
  "published_at": "2015-07-22T16:07:00+10:00",
  "created_at": "2014-07-26T22:31:14+10:00",
  "vendor": "Poison City Estore",
  "type": "",
  "tags": [
    "LP",
    "Vinyl"
  ],
  "price": 4400,
  "price_min": 4400,
  "price_max": 4400,
  "available": false,
  "price_varies": false,
  "compare_at_price": null,
  "compare_at_price_min": 0,
  "compare_at_price_max": 0,
  "compare_at_price_varies": false,
  "variants": [
    {
      "id": 757976871,
      "title": "Default Title",
      "option1": "Default Title",
      "option2": null,
      "option3": null,
      "sku": "",
      "requires_shipping": true,
      "taxable": true,
      "featured_image": null,
      "available": false,
      "name": "DESCENDENTS 'Everything Sucks' LP",
      "public_title": null,
      "options": [
        "Default Title"
      ],
      "price": 4400,
      "weight": 400,
      "compare_at_price": null,
      "inventory_quantity": 0,
      "inventory_management": "shopify",
      "inventory_policy": "deny",
      "barcode": "",
      "requires_selling_plan": false,
      "selling_plan_allocations": []
    }
  ],
  "images": [
    "//cdn.shopify.com/s/files/1/0586/2689/products/Descendents_Sucks.jpg?v=1406377874"
  ],
  "featured_image": "//cdn.shopify.com/s/files/1/0586/2689/products/Descendents_Sucks.jpg?v=1406377874",
  "options": [
    "Title"
  ],
  "media": [
    {
      "alt": null,
      "id": 3657302149,
      "position": 1,
      "preview_image": {
        "aspect_ratio": 1,
        "height": 500,
        "width": 500,
        "src": "https://cdn.shopify.com/s/files/1/0586/2689/products/Descendents_Sucks.jpg?v=1406377874"
      },
      "aspect_ratio": 1,
      "height": 500,
      "media_type": "image",
      "src": "https://cdn.shopify.com/s/files/1/0586/2689/products/Descendents_Sucks.jpg?v=1406377874",
      "width": 500
    }
  ],
  "requires_selling_plan": false,
  "selling_plan_groups": [],
  "content": "<span>BACK IN STOCK! On an undisclosed date in 1996, deemed by some as Science's Finest Hour, Milo hung up his lab coat and the Descendents began work on a new LP. Produced by Bill and Stephen at the band's own studio in Ft. Collins, CO (aptly dubbed The Blasting Room), \"Everything Sucks\" picks up EXACTLY where the Descendents left off. The tracks \"Everything Sucks,\" \"This Place\" and \"Coffee Mug\" are classic Descendents, raging along at the speed of sound, while \"I'm The One\", \"Sick-O-Me\" and \"She Loves Me\" are anthems of the \"girl-song\" milieu.</span>"
}
*/

func (a *PoisonCity) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(PC_SEARCH_URL, url.QueryEscape(artist))
	return query
}

func (a *PoisonCity) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

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
		if strings.Index(t, "row results") >= 0 {
			sku := SKU{}
			subUrl := strings.TrimPrefix(strings.TrimSpace(toks[idx+3]), "<a href=\"")
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

			sku.Image = fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"%s\">", sku.Image)
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
