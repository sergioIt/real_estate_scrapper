package main

import (
	"fmt"
	"log"
	"web_parser/pkg/scraper"
)

func main() {
	// Create scraper instance
	s := scraper.NewSSGEScraper()

	// Set up default filters to get flats for sale
	filters := scraper.FilterParams{
		PriceFrom:    0, // No minimum price
		PriceTo:      0, // No maximum price
		Type:         "For sale",
		PropertyType: "Flat",
	}

	log.Println("Starting scrape from home.ss.ge...")

	// Perform the scraping
	properties, err := s.Scrape(filters)
	if err != nil {
		log.Fatalf("Failed to scrape: %v", err)
	}

	// Limit to 10 records
	maxRecords := 10
	if len(properties) > maxRecords {
		properties = properties[:maxRecords]
	}

	// Output results to stdout
	fmt.Printf("\n=== Found %d unique properties ===\n\n", len(properties))

	for i, prop := range properties {
		fmt.Printf("%d. %s\n", i+1, prop.Title)
		fmt.Printf("   Price per m²: %s\n", prop.PricePerSqm)
		fmt.Printf("   Full Price: %s\n", prop.FullPrice)
		fmt.Printf("   Location: %s\n", prop.Location)
		if prop.SquareMeters > 0 {
			fmt.Printf("   Size: %d m²\n", prop.SquareMeters)
		}
		fmt.Printf("   URL: %s\n", prop.URL)
		fmt.Println()
	}

	fmt.Printf("Total: %d unique properties\n", len(properties))
}
