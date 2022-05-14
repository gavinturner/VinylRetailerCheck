package retailers

import "github.com/pkg/errors"

const (
	AF_PREFIX = "https://artistfirst.com.au"
	AF_URL    = "https://artistfirst.com.au/search?q=%s+vinyl"
)

type ArtistFirst struct{}

func (a *ArtistFirst) GetArtistQueryURL(artist string) string {
	query := fmt.Sprintf(AF_URL, url.QueryEscape(artist))
	return query
}

func (a *ArtistFirst) ScrapeArtistReleases(artist string, prices *KnownPrices) (findings []SKU, found bool) {

	findings = []SKU{}
	query := a.GetArtistQueryURL(artist)
	resp, err := http.Get(query)
	if err != nil {
		fmt.Printf("[FAILED: %s]..", err.Error())
		return "", false
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[FAILED: %s]..", err.Error())
		return "", false
	}
	toks := strings.Split(string(body), ">")

	for idx, t := range toks {
		if strings.Index(t, "product-list-item-title") >= 0 {

			sku := SKU{
				Url:    strings.Replace(toks[idx+1]+">", "<a href=\"", fmt.Sprintf("<a href=\"%s", AF_PREFIX), -1),
				Artist: strings.ToLower(strings.TrimSpace(strings.TrimSuffix(toks[idx-1], "</p"))),
				Image:  strings.TrimSpace(toks[idx-10]) + ">",
				Name:   strings.TrimSpace(strings.TrimSuffix(toks[idx+2], "</a")),
				Price:  toks[idx+5], // sold out
			}

			sku.Image = strings.Replace(sku.Image, "<img src=\"", fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"https:"), -1)
			if strings.Index(strings.ToLower(sku.Price), "sold out") < 0 {
				// ok we didnt find the sold out price
				sku.Price = toks[idx+6] // price?

				if strings.Index(strings.ToLower(sku.Price), "class=\"money") >= 0 {
					// ok we found the 'on special' case - one more..
					sku.Price = toks[idx+7] // price (when on sale)
					sku.Image = strings.TrimSpace(toks[idx-10])
				} else {
					sku.Image = strings.TrimSpace(toks[idx-8])
				}
				sku.Image += ">"
				sku.Image = strings.Replace(image, "<img src=\"", fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"https:"), -1)
				sku.Price = strings.TrimSuffix(price, "</span")
			} else {
				sku.Price = "sold out"
			}
			sku.Price = strings.TrimSpace(sku.Price)

			if sku.Artist != artist {
				continue
			}
			findings = append(findings, sku)
		}
	}
	return findings, found
}
