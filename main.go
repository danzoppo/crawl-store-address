// Package main obtains addresses of all CVS stores across the USA.
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

// storeStatesUrl is the URL to the different CVS state store pages.
const baseSearchURL = "https://www.cvs.com/store-locator/cvs-pharmacy-locations"

func main() {
	// Get today's date as store locations with change over time
	today := date.Today().String()

	// Setup csv file to store addresses
	fName := "cvs-store-locations-" + today + ".csv"
	file, err := os.Create(fName)
	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}
	defer file.Close()

	// Initialize writer and memory flush
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Set headers
	writer.Write([]string{"address"})

	// Instantiate collector. This collector will visit each of the state
	// webpages to find the stores in each state.
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

	// Clone collector to scrape that addresses.
	cityCollector := c.Clone()
	storeCollector := c.Clone()

	// Visit each of the states
	c.OnHTML("div.states li", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		link = e.Request.AbsoluteURL(link)
		// Start visiting each city within a state
		cityCollector.Visit(link)
	})

	// Notify the weblink for each state
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting State:", r.URL.String())
	})

	// Create a collector to acquire the store addresses.
	cityCollector.OnHTML("div.states li", func(e *colly.HTMLElement) {
		link := e.ChildAttr("a[href]", "href")
		link = e.Request.AbsoluteURL(link)
		storeCollector.Visit(link)
	})

	// Notify with the weblink for each city/town visited.
	cityCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting town:", r.URL.String())
	})

	// Obtain store address
	storeCollector.OnHTML("p.store-address", func(e *colly.HTMLElement) {
		address := strings.TrimSpace(e.Text)

		// Print to console to follow progress
		fmt.Println(address)
		// Write address to csv file
		writer.Write([]string{address})
	})

	// Commence searching from the base URL.
	startPage := c.Visit(baseSearchURL)
	// Check for errors
	fmt.Println(startPage)
}
