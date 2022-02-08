// Package main scrapes the addresses for every CVS Pharmacy
// from the CVS Pharmacy website.
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fxtlabs/date"
	"github.com/gocolly/colly/v2"
)

// baseSearchURL is the URL to the different CVS state store pages.
const baseSearchURL = "https://www.cvs.com/store-locator/cvs-pharmacy-locations"

func main() {
	// Get today's date as store locations with change over time.
	today := date.Today().String()

	// Setup csv file for data storage.
	fName := today + "_" + "cvs-store-locations" + ".csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Initialize writer and flush memory.
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Set header.
	writer.Write([]string{"address"})

	// Instantiate the collector. This collector will visit each of the state
	// and town webpages to find the addresses.
	c := colly.NewCollector(
		// Set allowed domains:
		colly.AllowedDomains("cvs.com", "www.cvs.com"),
	)

	// Set LimitRules to not bother the site.
	c.Limit(&colly.LimitRule{
		DomainGlob:  "cvs.com/*",
		Delay:       1 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	// Identity state and town html elements and visit links.
	c.OnHTML("div.states li", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		link = e.Request.AbsoluteURL(link)
		c.Visit(link)
	})

	// Print URL to track scraping.
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	// Obtain store address and write to csv file. Print
	// address to console to track scraping.
	c.OnHTML("p.store-address", func(e *colly.HTMLElement) {
		address := strings.TrimSpace(e.Text)
		fmt.Println(address)
		writer.Write([]string{address})
	})

	// Commence searching from the base URL.
	err = c.Visit(baseSearchURL)
	log.Fatal(err)
}
