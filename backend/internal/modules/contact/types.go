package contact

// ContactSubmit mirrors openapi.yaml schema ContactSubmit. `website` is the
// honeypot (must be empty); turnstile_token is the Cloudflare Turnstile token.
type ContactSubmit struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Subject        string `json:"subject"`
	Message        string `json:"message"`
	Locale         string `json:"locale"`
	Website        string `json:"website"`         // honeypot
	TurnstileToken string `json:"turnstile_token"` //
}

// ContactMessage mirrors openapi.yaml schema ContactMessage (admin inbox).
type ContactMessage struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Subject   *string `json:"subject"`
	Message   string  `json:"message"`
	Locale    *string `json:"locale"`
	IP        *string `json:"ip"`
	UserAgent *string `json:"user_agent"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// OkResult is the { ok: true } accepted body.
type OkResult struct {
	OK bool `json:"ok"`
}

// ContactPage is the paginated admin inbox listing.
type ContactPage struct {
	Items      []ContactMessage `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

// StatusUpdate is the PATCH body { status }.
type StatusUpdate struct {
	Status string `json:"status"`
}
