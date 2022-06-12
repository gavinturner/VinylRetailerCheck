package retailers

import (
	"encoding/csv"
	"github.com/gavinturner/vinylretailers/util/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// https://damagedmusic.com.au/?s=clowns+vinyl&post_type=product

const (
	BD_URL_PREFIX   = "https://beatdisc.com.au"
	DATA_DIR        = "./retailers/data"
	EMPTY_IMAGE_URL = "https://www.freeiconspng.com/thumbs/no-image-icon/no-image-icon-6.png"
)

type BeatDiscRecords struct{}

func (a *BeatDiscRecords) GetArtistQueryURL(artist string) string {
	return "n/a"
}

func (a *BeatDiscRecords) ScrapeArtistReleases(artist string) (findings []SKU, err error) {

	findings = []SKU{}
	findingsMap := map[string]SKU{}

	// find the latest beatdisc data file
	files, err := ioutil.ReadDir(DATA_DIR)
	if err != nil {
		log.Error(err, "Failed to list files in dir %s", DATA_DIR)
		return []SKU{}, errors.Wrapf(err, "failed to list files in dir %s", DATA_DIR)
	}

	dataFiles := []string{}
	for _, file := range files {
		if !file.IsDir() && strings.Index(file.Name(), "beatdisc") > 0 {
			dataFiles = append(dataFiles, file.Name())
		}
	}
	if len(dataFiles) == 0 {
		log.Errorf("No beatdisc stock file found in dir %s", DATA_DIR)
		return []SKU{}, errors.Wrapf(err, "no beatdisc stock file found in dir %s", DATA_DIR)
	}
	sort.Strings(dataFiles)
	path := DATA_DIR + "/" + dataFiles[len(dataFiles)-1]
	file, err := os.Open(path)
	if err != nil {
		log.Error(err, "Failed to open data file @ %s", path)
		return []SKU{}, errors.Wrapf(err, "failed to open data file @ %s", path)
	}
	defer file.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(file)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Error(err, "Failed to read csv data file @ %s", path)
		return []SKU{}, errors.Wrapf(err, "failed to read csv data file @ %s", path)
	}
	for i, line := range data {
		if i > 0 { // omit header line
			sku := SKU{
				Url:    "unavailable online",
				Artist: strings.ToLower(line[0]),
				Name:   line[1] + " " + line[8] + " [" + line[11] + "]",
				Price:  line[3],
				Image:  EMPTY_IMAGE_URL,
			}
			if sku.Artist != strings.ToLower(artist) {
				continue
			}
			sku.Price = sku.Price[2:]
			sku.Price = "$" + strings.TrimSpace(sku.Price)
			image, err := FindCoverURL(artist, line[1])
			if err != nil {
				log.Error(err, "Failed to get image for release")
				return []SKU{}, errors.Wrapf(err, "failed to get image for release")
			}
			if image != "" {
				sku.Image = image
			}
			// make sure we handel dupe titles
			findingsMap[sku.Name] = sku
		}
	}

	// now export the unduped titles
	for _, sku := range findingsMap {
		findings = append(findings, sku)
	}

	return findings, nil
}
