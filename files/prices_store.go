package files

import (
	"bufio"
	"fmt"
	"github.com/gavinturner/vinylretailers/db"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const (
	PRICES_STORE = "./cfg/af_known.txt"
)

// map of artist, map of release, price
type PricesStore map[string]map[string]string

// read our cfg file of known prices..
func (k *PricesStore) LoadFromFile(filename string) error {

	if k == nil {
		*k = PricesStore{}
	}
	dataFile, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open cfg file %s", filename)
	}
	if err == nil {
		scanner := bufio.NewScanner(dataFile)
		for scanner.Scan() {
			line := scanner.Text()
			toks := strings.Split(line, "|")
			if _, ok := (*k)[toks[0]]; !ok {
				(*k)[toks[0]] = map[string]string{}
			}
			(*k)[toks[0]][toks[1]] = toks[2]
		}
	}
	dataFile.Close()
	return nil
}

func (k *PricesStore) ApplyScrapings(artist string, findings []db.SKU) (newPrices []db.SKU) {

	if k == nil {
		*k = PricesStore{}
	}
	newPrices = []db.SKU{}

	for _, sku := range findings {
		// grab the existing price
		if _, ok := (*k)[artist]; !ok {
			(*k)[artist] = map[string]string{}
		}
		existingPrice := (*k)[artist][sku.Name]

		// does new price match existing price, and not sold out?
		if existingPrice != sku.Price && sku.Price != "sold out" {
			if existingPrice == "" {
				newPrices = append(newPrices, sku)
			} else {
				newPrices = append(newPrices, sku)
			}
		}
		// store new price
		(*k)[artist][sku.Name] = sku.Price
	}
	return newPrices
}

func (k *PricesStore) DumpToFile(filename string) error {
	// write the cfg file back out for next pass..
	if k == nil {
		return fmt.Errorf("no prices store exists")
	}
	dataFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open outfile %s", filename)
	}
	for artist := range *k {
		for name, price := range (*k)[artist] {
			dataFile.WriteString(fmt.Sprintf("%s|%s|%s\n", artist, name, price))
		}
	}
	if err := dataFile.Close(); err != nil {
		return errors.Wrapf(err, "failed to close outfile %s", filename)
	}
	return nil
}
