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
	UTOP_URL_PREFIX         = "https://utopia.com.au"
	UTOP_PRODUCT_URL_PREFIX = "https://utopia.com.au/collections/all/products"
	UTOP_SEARCH_INSTOCK_URL = "/search?q=%s+vinyl&_=pf&pf_st_currently_in_stock=true&pf_pt_product_type=New+Vinyl+-+12 Inch"
)

type Utopia struct{}

func (a *Utopia) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(UTOP_SEARCH_INSTOCK_URL, url.QueryEscape(artist))
	query = UTOP_URL_PREFIX + query
	return query
}

type UtopiaSKUEntry struct {
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
			AspectRatio float64 `json:"aspect_ratio"`
			Height      int     `json:"height"`
			Width       int     `json:"width"`
			Src         string  `json:"src"`
		} `json:"preview_image"`
		AspectRatio float64 `json:"aspect_ratio"`
		Height      int     `json:"height"`
		MediaType   string  `json:"media_type"`
		Src         string  `json:"src"`
		Width       int     `json:"width"`
	} `json:"media"`
	RequiresSellingPlan bool          `json:"requires_selling_plan"`
	SellingPlanGroups   []interface{} `json:"selling_plan_groups"`
	Content             string        `json:"content"`
}

func (a *Utopia) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	findings = []SKU{}

	query := a.GetArtistQueryURL(artist)
	resp, err := http.Get(query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve search query %s", query)
	}
	bodyInStock, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract body of search results %s", query)
	}

	toks := strings.Split(string(bodyInStock), ">")
	skus := []UtopiaSKUEntry{}
	for _, t := range toks {
		str := t
		idx := strings.Index(str, "collection.push(")
		for idx >= 0 {
			jd := str[idx+16:]
			jd = jd[0 : strings.Index(jd, "});")+1]
			str = str[idx+16:]
			idx = strings.Index(str, "collection.push(")
			data := UtopiaSKUEntry{}
			json.Unmarshal([]byte(jd), &data)
			skus = append(skus, data)
		}
	}
	// now go through skus and create findings
	for _, s := range skus {

		title := s.Title
		if strings.Index(title, " - Vinyl - New") > 0 {
			title = title[0:strings.Index(title, " - Vinyl - New")]
		}

		sku := SKU{
			Url:    UTOP_PRODUCT_URL_PREFIX + s.Handle,
			Artist: strings.ToLower(s.Vendor),
			Name:   title,
			Price:  fmt.Sprintf("$%.02f", float64(s.Price)/100.0),
			Image:  "<img width=\"150px\" height=\"150px\" src=\"https:" + s.FeaturedImage + "\" />",
		}

		if !s.Available {
			sku.Price = SOLD_OUT
		}

		// artist name can contain the searched for artist if it's a split etc.
		if strings.Index(sku.Artist, strings.ToLower(artist)) < 0 {
			continue
		}
		findings = append(findings, sku)
	}
	return findings, nil
}
