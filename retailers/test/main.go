package main

import (
	"bufio"
	"fmt"
	"github.com/gavinturner/vinylretailers/retailers"
	"os"
	"strings"
)

func main() {
	fmt.Printf("Scrape test\n")
	scraper := retailers.RepressedRecords{}

	// multi artist test
	readFile, err := os.Open("./cfg/artists.txt")
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

	skus := []retailers.SKU{}
	for _, artist := range artists {
		if len(artist) == 0 {
			continue
		}
		fmt.Printf("ARTIST: %s", artist)
		s, _ := scraper.ScrapeArtistReleases(artist)
		skus = append(skus, s...)
		fmt.Printf(" (%v)\n", len(s))
	}
	for _, s := range skus {
		fmt.Printf("SKU> %s, %s, (%s)\n", s.Artist, s.Name, s.Price)
		fmt.Printf("IMAGE> %s\n", s.Image)
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
