package media

// Media mirrors openapi.yaml schema Media.
type Media struct {
	ID        int64   `json:"id"`
	S3Key     string  `json:"s3_key"`
	PublicURL string  `json:"public_url"`
	Mime      string  `json:"mime"`
	Width     *int64  `json:"width"`
	Height    *int64  `json:"height"`
	SizeBytes int64   `json:"size_bytes"`
	AltCS     *string `json:"alt_cs"`
	AltEN     *string `json:"alt_en"`
	CreatedAt string  `json:"created_at"`
}

// MediaPage is the paginated admin media listing.
type MediaPage struct {
	Items      []Media `json:"items"`
	NextCursor *string `json:"next_cursor"`
}
