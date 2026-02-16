package scraper

// Property represents a real estate listing
type Property struct {
	ID           string  // Unique hash for duplicate detection
	Title        string  // e.g. "3 room Flat For Sale"
	Price        string  // Total price in USD
	PriceNumeric float64 // Numeric price for sorting/filtering
	PricePerSqm        string  // Price per square meter
	PricePerSqmNumeric float64 // Numeric price per m² (USD only)
	FullPrice    string  // Full total price (in thousands, e.g. "1,579 $")
	Location     string  // City area/district
	SquareMeters int     // Property size in m²
	URL          string
	Description  string
	AddedDate    string
	ListingType  string // "sale" or "rent"
}

// FilterParams represents the search filters for real estate
type FilterParams struct {
	PriceFrom    int
	PriceTo      int
	Type         string // For sale, For rent
	PropertyType string // Apartment, House, etc.
	City         string
	District     string
	PricePerSqmFrom int // Min price per m² in USD (0 = no bound)
	PricePerSqmTo   int // Max price per m² in USD (0 = no bound)
	AreaFrom        int // Min area in m² (0 = no bound)
	AreaTo          int // Max area in m² (0 = no bound)
	Limit           int // Max number of results to return (0 = default 10)
}

// RealEstateScraper defines the interface for real estate website scrapers
type RealEstateScraper interface {
	// Scrape performs the scraping operation with given filters
	Scrape(filters FilterParams) ([]Property, error)
}
