package files

import "github.com/pkg/errors"

const (
	OUT_FILE = "./af_known.txt"
)

// map of artist, map of release, price
type PricesStore map[string]map[string]string

func (k *PricesStore) LoadFromFile(filename string) error {

	// read our data file of known prices..
	dataFile, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open data file %s", filename)
	}
	if err == nil {
		scanner := bufio.NewScanner(dataFile)
		for scanner.Scan() {
			line := scanner.Text()
			toks := strings.Split(line, "|")
			if _, ok := data[toks[0]]; !ok {
				k[toks[0]] = map[string]string{}
			}
			k[toks[0]][toks[1]] = toks[2]
		}
	}
	dataFile.Close()
	return nil
}

func (k *PricesStore) ApplyScrapings(artist string, findings []SKU) (newPrices []SKU) {

	newPrices = []SKU{}

	for _, sku := range findings {
		// grab the existing price
		if _, ok := k[artist]; !ok {
			k[artist] = map[string]string{}
		}
		existingPrice := k[artist][sku.Name]

		// does new price match existing price, and not sold out?
		if existingPrice != sku.Price && sku.Price != "sold out" {
			if existingPrice == "" {
				newPrices = append(newPrices, sku)
				findings += renderResultRow(true, image, artist, url, name, price, existingPrice)
				//fmt.Printf("NEW PRODUCT DETECTED: %s %s, %s%s</a>, %s\n", image, artist, url, name, price)
			} else {
				newPrices = append(newPrices, sku)
				findings += renderResultRow(false, image, artist, url, name, price, existingPrice)
				//fmt.Printf("NEW PRICE DETECTED: %s %s, %s%s</a>, %s (%s)\n", image, artist, url, name, price, existingPrice)
			}
			found = true
		}
		// store new price
		k[artist][name] = price
	}
	return newPrices
}

func (k *PricesStore) DumpToFile(filename string) error {
	// write the data file back out for next pass..
	dataFile, err = os.Create(filename)
	if err != nil {
		return errors.Wrapf("failed to open outfile %s", filename)
	}
	for artist := range k {
		for name, price := range k[artist] {
			dataFile.WriteString(fmt.Sprintf("%s|%s|%s\n", artist, name, price))
		}
	}
	if err := dataFile.Close(); err != nil {
		return errors.Wrapf("failed to close outfile %s", filename)
	}
	return nil
}
