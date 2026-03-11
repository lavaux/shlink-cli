// Package api defines the data types returned by the Shlink REST API.
package api

// ---- Short URLs ----

// ShortURL represents a Shlink short URL resource.
type ShortURL struct {
	ShortCode     string        `json:"shortCode"`
	ShortURL      string        `json:"shortUrl"`
	LongURL       string        `json:"longUrl"`
	DateCreated   string        `json:"dateCreated"`
	Tags          []string      `json:"tags"`
	Meta          ShortURLMeta  `json:"meta"`
	Domain        *string       `json:"domain"`
	Title         *string       `json:"title"`
	Crawlable     bool          `json:"crawlable"`
	ForwardQuery  bool          `json:"forwardQuery"`
	VisitsSummary VisitsSummary `json:"visitsSummary"`
}

// VisitsSummary holds visit counts for a short URL.
type VisitsSummary struct {
	Total   int `json:"total"`
	NonBots int `json:"nonBots"`
	Bots    int `json:"bots"`
}

// ShortURLMeta holds optional metadata for a short URL.
type ShortURLMeta struct {
	ValidSince *string `json:"validSince"`
	ValidUntil *string `json:"validUntil"`
	MaxVisits  *int    `json:"maxVisits"`
}

// ShortURLList is the paginated list response for short URLs.
type ShortURLList struct {
	ShortURLs ShortURLPage `json:"shortUrls"`
}

// ShortURLPage holds the paginated short URL data.
type ShortURLPage struct {
	Data       []ShortURL `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds paging metadata.
type Pagination struct {
	CurrentPage        int `json:"currentPage"`
	PagesCount         int `json:"pagesCount"`
	ItemsPerPage       int `json:"itemsPerPage"`
	ItemsInCurrentPage int `json:"itemsInCurrentPage"`
	TotalItems         int `json:"totalItems"`
}

// CreateShortURLRequest is the body for creating a short URL.
type CreateShortURLRequest struct {
	LongURL      string   `json:"longUrl"`
	ShortCode    string   `json:"customSlug,omitempty"`
	Title        string   `json:"title,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Domain       string   `json:"domain,omitempty"`
	MaxVisits    *int     `json:"maxVisits,omitempty"`
	ValidSince   string   `json:"validSince,omitempty"`
	ValidUntil   string   `json:"validUntil,omitempty"`
	Crawlable    *bool    `json:"crawlable,omitempty"`
	ForwardQuery *bool    `json:"forwardQuery,omitempty"`
	FindIfExists bool     `json:"findIfExists,omitempty"`
}

// EditShortURLRequest is the body for editing a short URL.
type EditShortURLRequest struct {
	LongURL      string   `json:"longUrl,omitempty"`
	Title        *string  `json:"title,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	MaxVisits    *int     `json:"maxVisits,omitempty"`
	ValidSince   *string  `json:"validSince,omitempty"`
	ValidUntil   *string  `json:"validUntil,omitempty"`
	Crawlable    *bool    `json:"crawlable,omitempty"`
	ForwardQuery *bool    `json:"forwardQuery,omitempty"`
}

// ---- Tags ----

// TagList is the response for listing tags.
type TagList struct {
	Tags TagData `json:"tags"`
}

// TagData holds the tag names.
type TagData struct {
	Data []string `json:"data"`
}

// TagStatsList is the response for listing tag stats.
type TagStatsList struct {
	Tags TagStatsData `json:"tags"`
}

// TagStatsData holds tag stat entries.
type TagStatsData struct {
	Data       []TagStat  `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// TagStat holds statistics for a single tag.
type TagStat struct {
	Tag            string        `json:"tag"`
	ShortURLsCount int           `json:"shortUrlsCount"`
	VisitsSummary  VisitsSummary `json:"visitsSummary"`
}

// RenameTagRequest renames a tag.
type RenameTagRequest struct {
	OldName string `json:"oldName"`
	NewName string `json:"newName"`
}

// ---- Visits ----

// VisitList is the paginated list of visits.
type VisitList struct {
	Visits VisitPage `json:"visits"`
}

// VisitPage holds paginated visit data.
type VisitPage struct {
	Data       []Visit    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Visit represents a single visit record.
type Visit struct {
	Referer       string    `json:"referer"`
	Date          string    `json:"date"`
	UserAgent     string    `json:"userAgent"`
	VisitLocation *Location `json:"visitLocation"`
	PotentialBot  bool      `json:"potentialBot"`
	VisitedURL    string    `json:"visitedUrl"`
	Type          string    `json:"type"`
}

// Location holds geographic visit data.
type Location struct {
	CountryName string `json:"countryName"`
	RegionName  string `json:"regionName"`
	CityName    string `json:"cityName"`
}

// GlobalVisits is the response for /visits.
type GlobalVisits struct {
	Visits GlobalVisitCounts `json:"visits"`
}

// GlobalVisitCounts holds the global visit counters.
type GlobalVisitCounts struct {
	NonOrphanVisits VisitsSummary `json:"nonOrphanVisits"`
	OrphanVisits    VisitsSummary `json:"orphanVisits"`
}

// ---- Domains ----

// DomainList is the response for listing domains.
type DomainList struct {
	Domains DomainData `json:"domains"`
}

// DomainData holds domain entries and a default redirect set.
type DomainData struct {
	Data             []Domain         `json:"data"`
	DefaultRedirects *DomainRedirects `json:"defaultRedirects"`
}

// Domain represents a configured domain.
type Domain struct {
	Domain    string           `json:"domain"`
	IsDefault bool             `json:"isDefault"`
	Redirects *DomainRedirects `json:"redirects"`
}

// DomainRedirects holds the three redirect URLs for a domain.
type DomainRedirects struct {
	BaseUrlRedirect    *string `json:"baseUrlRedirect"`
	Regular404Redirect *string `json:"regular404Redirect"`
	InvalidShortURL    *string `json:"invalidShortUrlRedirect"`
}

// ---- Health ----

// Health is the response from /rest/health.
type Health struct {
	Status  string `json:"status"`
	Version string `json:"version"`
	Links   struct {
		About   string `json:"about"`
		Project string `json:"project"`
	} `json:"links"`
}

// ---- Redirect Rules ----

// RedirectRuleSet holds the redirect rules for a short URL.
type RedirectRuleSet struct {
	RedirectRules []RedirectRule `json:"redirectRules"`
}

// RedirectRule defines a single redirect rule.
type RedirectRule struct {
	LongURL    string              `json:"longUrl"`
	Priority   int                 `json:"priority"`
	Conditions []RedirectCondition `json:"conditions"`
}

// RedirectCondition is a single condition in a redirect rule.
type RedirectCondition struct {
	Type       string `json:"type"`
	MatchKey   string `json:"matchKey"`
	MatchValue string `json:"matchValue"`
}
