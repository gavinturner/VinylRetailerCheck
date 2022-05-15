// artist_first project main.go
package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gavinturner/VinylRetailChecker/files"
	gomail "gopkg.in/mail.v2"
)

const (
	CITADEL_PREFIX = "https://www.citadelmailorder.com"
	CITADEL_URL    = "https://www.citadelmailorder.com/jsp/cmos/merch/hard_ons_merch_page.jsp?source=hard_ons_merch_page_new"
	PC_PREFIX      = "https://poisoncityestore.com"
	PC_URL         = "https://poisoncityestore.com/search?q=%s+vinyl"
	PERIOD_MINS    = 15
)

func main() {

	fmt.Println("Vinyl checker")
	data := files.PricesStore{}
	err := data.LoadFromFile(files.PRICES_STORE)
	if err != nil {
		panic(err)
	}

	// get the artists file path as cmd line param
	argsWithoutProg := os.Args[1:]
	fileName := ""
	if len(argsWithoutProg) >= 1 {
		fileName = argsWithoutProg[0]
	}
	if fileName == "" {
		fmt.Printf("No artists list supplied\n")
		os.Exit(1)
	}

	//
	// Check Artists First every 15 minutes
	//

	for {
		// read the supplied list of artists (one per line, first CLP)
		artists := ArtistsList
		err = artists.Read(fileName)
		if err != nil {
			panic(err)
		}

		// now scrape Artists First and find products from the artist search, noting things we haven seen before
		found := false
		message := "<table>"
		for _, artist := range artists {
			var str string
			artist = strings.TrimSpace(artist)
			str, fnd := processArtistsFirst(artist, data)
			message += str
			found = found || fnd
			str, fnd = processPoisonCity(artist, data)
			message += str
			found = found || fnd
		}
		message += "</table>"

		if found {
			err = sendEmail("gturner.au@gmail.com", "New Vinyl postings found", message)
			if err != nil {
				panic(err)
			}
			fmt.Printf("New vinyl postings found\n")
		} else {
			fmt.Printf("No new vinyl postings found\n")
		}

		// write the data file back out for next pass..
		err = data.DumpToFile(files.PRICES_STORE)
		if err != nil {
			panic(err)
		}
		// check out for our delay
		time.Sleep(PERIOD_MINS * 60 * time.Second)
	}

}

func processArtistsFirst(artist string, data map[string]map[string]string) (findings string, found bool) {
	fmt.Printf("Checking ArtistsFirst: %s...", artist)
	findings = ""

	query := fmt.Sprintf(AF_URL, url.QueryEscape(artist))
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
			url := strings.Replace(toks[idx+1]+">", "<a href=\"", fmt.Sprintf("<a href=\"%s", AF_PREFIX), -1)
			disc_artist := strings.ToLower(strings.TrimSpace(strings.TrimSuffix(toks[idx-1], "</p")))
			image := strings.TrimSpace(toks[idx-10]) + ">"
			image = strings.Replace(image, "<img src=\"", fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"https:"), -1)
			name := strings.TrimSpace(strings.TrimSuffix(toks[idx+2], "</a"))
			price := toks[idx+5] // sold out
			if strings.Index(strings.ToLower(price), "sold out") < 0 {
				// ok we didnt find the sold out price
				price = toks[idx+6] // price?

				if strings.Index(strings.ToLower(price), "class=\"money") >= 0 {
					// ok we found the 'on special' case - one more..
					price = toks[idx+7] // price (when on sale)
					image = strings.TrimSpace(toks[idx-10])
				} else {
					image = strings.TrimSpace(toks[idx-8])
				}
				image += ">"
				image = strings.Replace(image, "<img src=\"", fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"https:"), -1)
				price = strings.TrimSuffix(price, "</span")
				//fmt.Printf("\nNAME:[%s]: PRICE:[%s] IMAGE:[%s]\n", name, price, image)
			} else {
				price = "sold out"
			}
			price = strings.TrimSpace(price)
			//fmt.Printf("[%s][%s]: [%s][%s]\n", artist, name, price, url)

			if disc_artist != artist {
				continue
			}

			// grab the existing price
			if _, ok := data[artist]; !ok {
				data[artist] = map[string]string{}
			}
			existingPrice := data[artist][name]

			// does new price match existing price, and not sold out?
			if existingPrice != price && price != "sold out" {
				if existingPrice == "" {
					findings += renderResultRow(true, image, artist, url, name, price, existingPrice)
					//fmt.Printf("NEW PRODUCT DETECTED: %s %s, %s%s</a>, %s\n", image, artist, url, name, price)
				} else {
					findings += renderResultRow(false, image, artist, url, name, price, existingPrice)
					//fmt.Printf("NEW PRICE DETECTED: %s %s, %s%s</a>, %s (%s)\n", image, artist, url, name, price, existingPrice)
				}
				found = true
			}
			// store new price
			data[artist][name] = price
		}
	}
	fmt.Printf("done\n")
	return findings, found
}

func processPoisonCity(artist string, data map[string]map[string]string) (findings string, found bool) {
	fmt.Printf("Checking Poison City: %s...", artist)
	findings = ""
	query := fmt.Sprintf(PC_URL, url.QueryEscape(artist))
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
		if strings.Index(t, "row results") >= 0 {
			subUrl := strings.TrimPrefix(strings.TrimSpace(toks[idx+3]), "<a href=\"")
			subUrl = PC_PREFIX + strings.TrimSuffix(subUrl, "\"")
			image := ""
			name := ""
			price := ""
			//fmt.Printf("\nURL:[%s]\n", subUrl)

			// now read THAT url...
			query := fmt.Sprintf(subUrl)
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

			subToks := strings.Split(string(body), ">")
			for idx, tok := range subToks {
				if strings.Index(tok, "itemprop=\"image\"") > 0 {
					image = "https:" + strings.TrimPrefix(strings.TrimSpace(tok), "<meta itemprop=\"image\" content=\"")
					image = strings.TrimSuffix(image, "\" /")
					//fmt.Printf("IMAGE:[%s]\n", image)
				}
				if strings.Index(tok, "itemprop=\"name\"") >= 0 {
					name = strings.TrimSpace(subToks[idx+1])
					name = strings.TrimSuffix(name, "</h1")
					//fmt.Printf("NAME:[%s]\n", name)
				}
				if strings.Index(tok, "itemprop=\"price\"") >= 0 {
					price = strings.TrimSpace(subToks[idx+1])
					price = strings.TrimSuffix(price, "</span")
					//fmt.Printf("PRICE:[%s]\n", price)
				}
				if strings.Index(tok, "\"Sold Out\"") >= 0 {
					price = "sold out"
				}
			}

			if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(artist)) {
				continue
			}

			image = fmt.Sprintf("<img width=\"150px\" height=\"150px\" src=\"%s\">", image)

			// grab the existing price
			if _, ok := data[artist]; !ok {
				data[artist] = map[string]string{}
			}
			existingPrice := data[artist][name]
			if existingPrice != price {
				findings += renderResultRow(true, image, artist, fmt.Sprintf("<a href=\"%s\">", subUrl), name, price, existingPrice)
				found = true
				//fmt.Printf("FOUND: [%s | %s | %s]\n", name, image, price)
			}

			// store new price
			data[artist][name] = price
		}
	}
	fmt.Printf("done\n")
	return findings, found
}

func sendEmail(toAddress string, subject string, message string) error {
	m := gomail.NewMessage()

	// Set E-Mail sender
	m.SetHeader("From", "gturner.au@gmail.com")

	// Set E-Mail receivers
	m.SetHeader("To", toAddress)

	// Set E-Mail subject
	m.SetHeader("Subject", subject)

	// Set E-Mail body. You can set plain text or html with text/html
	m.SetBody("text/html", message)

	// Settings for SMTP server
	d := gomail.NewDialer("smtp.gmail.com", 587, "gturner.au@gmail.com", "exlxvgauubmdugzy")

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Now send E-Mail
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func renderResultRow(new bool, image string, artist string, url string, name string, price string, existingPrice string) string {
	htmlOut := "<tr>\n"
	htmlOut += fmt.Sprintf("<td>%s</td>\n", image)
	htmlOut += fmt.Sprintf("<td>%s<br>%s%s</a><br>%s</td>\n", artist, url, name, price)
	htmlOut += "</tr>\n"
	return htmlOut
}
