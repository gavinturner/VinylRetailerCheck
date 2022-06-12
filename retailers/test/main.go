package main

import (
	"bufio"
	"fmt"
	"github.com/gavinturner/vinylretailers/retailers"
	_ "github.com/lib/pq"
	"os"
	"strings"
)

func main() {
	fmt.Printf("Scrape test\n")
	scraper := retailers.MusicFarmers{}

	// multi artist test
	readFile, err := os.Open("./retailers/test/artists.txt")
	if err != nil {
		panic(err)
	}
	artists := []string{}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		artists = append(artists, strings.TrimSpace(strings.ToLower(fileScanner.Text())))
	}
	readFile.Close()

	artists = []string{"clowns"}

	skus := []retailers.SKU{}
	for _, artist := range artists {
		if len(artist) == 0 {
			continue
		}
		fmt.Printf("ARTIST: %s", artist)
		s, err := scraper.ScrapeArtistReleases(artist)
		if err != nil {
			fmt.Printf("ERROR: %s", err.Error())
			os.Exit(1)
		}
		skus = append(skus, s...)
		fmt.Printf(" (%v)\n", len(s))
	}
	for _, s := range skus {
		fmt.Printf("SKU> %s, %s, (%s)\n", s.Artist, s.Name, s.Price)
	}

	// single artist test
	/*
		sskus := []retailers.SKU{}
		artist := "clowns"
		fmt.Printf("\nARTIST: %s", artist)
		s, _ := scraper.ScrapeArtistReleases(artist)
		sskus = append(sskus, s...)
		fmt.Printf(" (%v)\n", len(s))

		for _, s := range sskus {
			fmt.Printf("SKU> [%s], [%s] (%s)\n", s.Artist, s.Name, s.Price)
			fmt.Printf("IMAGE: %s\n\n", s.Image)
		}
	*/
}
