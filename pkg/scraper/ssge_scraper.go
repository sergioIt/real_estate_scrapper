package scraper

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// SSGEScraper handles the scraping of ss.ge
type SSGEScraper struct {
	collector *colly.Collector
}

// NewSSGEScraper creates a new instance of SSGEScraper
func NewSSGEScraper() *SSGEScraper {
	c := colly.NewCollector(
		colly.AllowedDomains("home.ss.ge", "www.home.ss.ge", "ss.ge"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	return &SSGEScraper{
		collector: c,
	}
}

// Scrape performs the scraping operation with given filters
func (s *SSGEScraper) Scrape(filters FilterParams) ([]Property, error) {
	var properties []Property

	// Track seen URLs to avoid duplicates
	seenURLs := make(map[string]bool)

	// Create a fresh collector for each scrape to avoid duplicate handlers
	c := colly.NewCollector(
		colly.AllowedDomains("home.ss.ge", "www.home.ss.ge", "ss.ge"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
	)

	// Set up the collector callbacks
	// Look for links that contain property info (text includes price and title)
	c.OnHTML("a[href*='/real-estate/'][href*='-room-']", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		url := e.Request.AbsoluteURL(href)

		// Skip if we've already seen this URL
		if seenURLs[url] {
			return
		}

		// Skip if it's a filter link (contains query parameters) or photo links
		if strings.Contains(url, "?cityIdList") ||
			strings.Contains(url, "&subdistrictIds") ||
			strings.Contains(e.Text, "All photos") {
			return
		}

		// Extract text from the link
		text := e.Text
		if text == "" || len(text) < 10 {
			return
		}

		// Parse the text - pattern: "97,300 $m² - 1,160 $3 room Flat For Sale. Ortachala"
		// We need: total price (1,160 $), title (3 room Flat For Sale), location (Ortachala), sqm

		var title, location, price, pricePerSqm, fullPrice string
		var sqm int
		var pricePerSqmNumeric float64

		// Split text into lines
		lines := strings.Split(text, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || len(line) < 15 {
				continue
			}

			// Look for the main info line with price and title
			// Pattern: "FULL_PRICE $m² - PRICE_PER_SQM $ TITLE. LOCATION"
			// Example: "77,000 $m² - 1,674 $2 room Flat For Sale. Batumi"
			if strings.Contains(line, "m²") && strings.Contains(line, "room") {
				// Extract full price (first number before m²)
				// Pattern: "77,000 $m²" or "68,000 ₾m²"
				reFullPriceDollar := regexp.MustCompile(`([\d,]+)\s*\$m²`)
				reFullPriceLari := regexp.MustCompile(`([\d,]+)\s*₾m²`)

				if match := reFullPriceDollar.FindStringSubmatch(line); len(match) > 1 {
					fullPrice = match[1] + " $"
					price = fullPrice // Keep price for backward compatibility
				} else if match := reFullPriceLari.FindStringSubmatch(line); len(match) > 1 {
					fullPrice = match[1] + " ₾"
					price = fullPrice
				}

				// Extract square meters - it's usually at the end like "83.9 m²" or "60 m²"
				// Look for the LAST occurrence of number + m²
				reM2 := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*m²`)
				matches := reM2.FindAllStringSubmatch(text, -1)
				if len(matches) > 0 {
					// Take the last match (actual size, not price per m²)
					lastMatch := matches[len(matches)-1]
					if len(lastMatch) > 1 {
						sqmFloat, _ := strconv.ParseFloat(lastMatch[1], 64)
						sqm = int(sqmFloat)
					}
				}

				// Extract price per m² and title (after the m² - dash part)
				// Split by m² to get the part after
				parts := strings.Split(line, "m²")
				if len(parts) > 1 {
					afterM2 := parts[1]

					// Find price per m² with $ or ₾
					// Match pattern like "- 1,674 $" or "- 97,300 ₾"
					reDollar := regexp.MustCompile(`-\s*([\d,]+)\s*\$`)
					reLari := regexp.MustCompile(`-\s*([\d,]+)\s*₾`)

					if match := reDollar.FindStringSubmatch(afterM2); len(match) > 1 {
						pricePerSqm = match[1] + " $/m²"
						numStr := strings.ReplaceAll(match[1], ",", "")
						pricePerSqmNumeric, _ = strconv.ParseFloat(numStr, 64)
					} else if match := reLari.FindStringSubmatch(afterM2); len(match) > 1 {
						pricePerSqm = match[1] + " ₾/m²"
					}

					// Extract title - pattern like "3 room Flat For Sale. Location"
					// Look for pattern with digits followed by "room"
					reTitle := regexp.MustCompile(`(\d+\s+room\s+[^.]+)\.?\s*(.*)`)
					if match := reTitle.FindStringSubmatch(afterM2); len(match) > 1 {
						title = strings.TrimSpace(match[1])
						if len(match) > 2 {
							location = strings.TrimSpace(match[2])
							// Clean location from extra newlines
							if idx := strings.Index(location, "\n"); idx > 0 {
								location = location[:idx]
							}
						}
					}
				}

				// If we found the data, break
				if price != "" && title != "" {
					break
				}
			}
		}

		// Only add if we have title and price
		if title != "" && price != "" {
			seenURLs[url] = true

			property := Property{
				ID:                 generatePropertyID(title, price, location),
				Title:              title,
				Price:              price,
				PricePerSqm:        pricePerSqm,
				PricePerSqmNumeric: pricePerSqmNumeric,
				FullPrice:          fullPrice,
				Location:           location,
				SquareMeters:       sqm,
				URL:                url,
				ListingType:        filters.Type,
			}

			properties = append(properties, property)
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error while scraping: %v", err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Printf("Visiting: %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Finished: %s", r.Request.URL.String())
	})

	// Visit just the first page
	url := s.constructURL(filters)
	err := c.Visit(url)
	if err != nil {
		return nil, fmt.Errorf("failed to scrape ss.ge: %w", err)
	}

	// Apply limit
	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	if len(properties) > limit {
		properties = properties[:limit]
	}

	return properties, nil
}

// constructURL builds the URL with the filter parameters
func (s *SSGEScraper) constructURL(filters FilterParams) string {
	// Base URL for home.ss.ge with English locale
	baseURL := "https://home.ss.ge/en/real-estate/l/"

	// Determine property type path segment
	propertyType := "Flat" // Default to Flat/Apartment
	if filters.PropertyType == "House" {
		propertyType = "House"
	} else if filters.PropertyType == "Commercial" {
		propertyType = "Commercial"
	}

	// Determine listing type (sale or rent)
	listingType := "For-Sale"
	if filters.Type == "For rent" || filters.Type == "rent" {
		listingType = "For-Rent"
	}

	// Construct base path: /en/real-estate/l/Flat/For-Sale
	url := baseURL + propertyType + "/" + listingType

	// Add query parameters
	params := []string{}

	if filters.PricePerSqmFrom > 0 || filters.PricePerSqmTo > 0 {
		// priceType=2 means "per m²", currencyId=2 means USD
		params = append(params, "currencyId=2", "priceType=2")
		if filters.PricePerSqmFrom > 0 {
			params = append(params, fmt.Sprintf("priceFrom=%d", filters.PricePerSqmFrom))
		}
		if filters.PricePerSqmTo > 0 {
			params = append(params, fmt.Sprintf("priceTo=%d", filters.PricePerSqmTo))
		}
	} else {
		if filters.PriceFrom > 0 {
			params = append(params, fmt.Sprintf("priceFrom=%d", filters.PriceFrom))
		}
		if filters.PriceTo > 0 {
			params = append(params, fmt.Sprintf("priceTo=%d", filters.PriceTo))
		}
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return url
}

// generatePropertyID creates a unique hash for duplicate detection
func generatePropertyID(title, price, location string) string {
	// Normalize the input by trimming spaces and converting to lowercase
	normalized := strings.ToLower(strings.TrimSpace(title + price + location))

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(normalized))

	// Convert to hex string and return first 16 characters
	return fmt.Sprintf("%x", hash)[:16]
}
