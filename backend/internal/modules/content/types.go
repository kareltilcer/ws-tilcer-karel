package content

import "encoding/json"

// The DTOs mirror openapi.yaml exactly (field names + nullability). Nullable
// output fields use *T so they serialize as JSON null; required fields are plain.
// On PATCH, a nil pointer means "field absent → leave unchanged"; OptionalNullID
// distinguishes absent / null / value for the nullable FK fields.

// OptionalNullID captures a JSON integer field that may be absent, null, or a
// number — needed to tell "clear the cover" (null) from "leave it" (absent) on a
// PATCH. Present is false when the key was omitted entirely.
type OptionalNullID struct {
	Present bool
	Value   *int64 // nil when explicitly null
}

func (o *OptionalNullID) UnmarshalJSON(b []byte) error {
	o.Present = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var n int64
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	o.Value = &n
	return nil
}

// ---- media (shared shapes; the media module owns upload/delete) ----

// MediaRef is the lightweight media reference embedded in content responses.
type MediaRef struct {
	ID        int64   `json:"id"`
	PublicURL string  `json:"public_url"`
	AltCS     *string `json:"alt_cs"`
	AltEN     *string `json:"alt_en"`
	Width     *int64  `json:"width"`
	Height    *int64  `json:"height"`
}

// ---- category ----

type Category struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	NameCS    string `json:"name_cs"`
	NameEN    string `json:"name_en"`
	SortOrder int64  `json:"sort_order"`
	Visible   bool   `json:"visible"`
}

type CategoryCreate struct {
	Slug      string `json:"slug"`
	NameCS    string `json:"name_cs"`
	NameEN    string `json:"name_en"`
	SortOrder *int64 `json:"sort_order"`
	Visible   *bool  `json:"visible"`
}

type CategoryUpdate struct {
	Slug      *string `json:"slug"`
	NameCS    *string `json:"name_cs"`
	NameEN    *string `json:"name_en"`
	SortOrder *int64  `json:"sort_order"`
	Visible   *bool   `json:"visible"`
}

// ---- project ----

// ProjectLink is one out-link on a project (repo/live/demo).
type ProjectLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
	Type  string `json:"type,omitempty"`
}

type ProjectSummary struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Category    Category  `json:"category"`
	TitleCS     string    `json:"title_cs"`
	TitleEN     string    `json:"title_en"`
	SummaryCS   *string   `json:"summary_cs"`
	SummaryEN   *string   `json:"summary_en"`
	Cover       *MediaRef `json:"cover"`
	Featured    bool      `json:"featured"`
	Status      string    `json:"status"`
	ProjectDate *string   `json:"project_date"`
	SortOrder   int64     `json:"sort_order"`
}

type ProjectDetail struct {
	ProjectSummary
	BodyCS    *string       `json:"body_cs"`
	BodyEN    *string       `json:"body_en"`
	Gallery   []MediaRef    `json:"gallery"`
	Links     []ProjectLink `json:"links"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

type ProjectCreate struct {
	Slug         string        `json:"slug"`
	CategoryID   int64         `json:"category_id"`
	TitleCS      string        `json:"title_cs"`
	TitleEN      string        `json:"title_en"`
	SummaryCS    string        `json:"summary_cs"`
	SummaryEN    string        `json:"summary_en"`
	BodyCS       string        `json:"body_cs"`
	BodyEN       string        `json:"body_en"`
	CoverMediaID *int64        `json:"cover_media_id"`
	Gallery      []int64       `json:"gallery"`
	Links        []ProjectLink `json:"links"`
	ProjectDate  *string       `json:"project_date"`
	Featured     *bool         `json:"featured"`
	Status       *string       `json:"status"`
	SortOrder    *int64        `json:"sort_order"`
}

type ProjectUpdate struct {
	Slug         *string        `json:"slug"`
	CategoryID   *int64         `json:"category_id"`
	TitleCS      *string        `json:"title_cs"`
	TitleEN      *string        `json:"title_en"`
	SummaryCS    *string        `json:"summary_cs"`
	SummaryEN    *string        `json:"summary_en"`
	BodyCS       *string        `json:"body_cs"`
	BodyEN       *string        `json:"body_en"`
	CoverMediaID OptionalNullID `json:"cover_media_id"`
	Gallery      *[]int64       `json:"gallery"`
	Links        *[]ProjectLink `json:"links"`
	ProjectDate  *string        `json:"project_date"`
	Featured     *bool          `json:"featured"`
	Status       *string        `json:"status"`
	SortOrder    *int64         `json:"sort_order"`
}

// ---- link ----

type Link struct {
	ID         int64   `json:"id"`
	LabelCS    string  `json:"label_cs"`
	LabelEN    string  `json:"label_en"`
	URL        string  `json:"url"`
	Icon       *string `json:"icon"`
	Visible    bool    `json:"visible"`
	SortOrder  int64   `json:"sort_order"`
	ClickCount int64   `json:"click_count"`
}

type LinkCreate struct {
	LabelCS   string  `json:"label_cs"`
	LabelEN   string  `json:"label_en"`
	URL       string  `json:"url"`
	Icon      *string `json:"icon"`
	Visible   *bool   `json:"visible"`
	SortOrder *int64  `json:"sort_order"`
}

type LinkUpdate struct {
	LabelCS   *string `json:"label_cs"`
	LabelEN   *string `json:"label_en"`
	URL       *string `json:"url"`
	Icon      *string `json:"icon"`
	Visible   *bool   `json:"visible"`
	SortOrder *int64  `json:"sort_order"`
}

// ---- page section ----

type Section struct {
	Key       string    `json:"key"`
	HeadingCS *string   `json:"heading_cs"`
	HeadingEN *string   `json:"heading_en"`
	BodyCS    *string   `json:"body_cs"`
	BodyEN    *string   `json:"body_en"`
	Media     *MediaRef `json:"media"`
	UpdatedAt string    `json:"updated_at"`
}

type SectionUpdate struct {
	HeadingCS *string        `json:"heading_cs"`
	HeadingEN *string        `json:"heading_en"`
	BodyCS    *string        `json:"body_cs"`
	BodyEN    *string        `json:"body_en"`
	MediaID   OptionalNullID `json:"media_id"`
}

// ---- skill ----

type Skill struct {
	ID        int64   `json:"id"`
	NameCS    string  `json:"name_cs"`
	NameEN    string  `json:"name_en"`
	Category  string  `json:"category"`
	Icon      *string `json:"icon"`
	Level     *int64  `json:"level"`
	Visible   bool    `json:"visible"`
	SortOrder int64   `json:"sort_order"`
}

type SkillCreate struct {
	NameCS    string  `json:"name_cs"`
	NameEN    string  `json:"name_en"`
	Category  string  `json:"category"`
	Icon      *string `json:"icon"`
	Level     *int64  `json:"level"`
	Visible   *bool   `json:"visible"`
	SortOrder *int64  `json:"sort_order"`
}

type SkillUpdate struct {
	NameCS    *string        `json:"name_cs"`
	NameEN    *string        `json:"name_en"`
	Category  *string        `json:"category"`
	Icon      *string        `json:"icon"`
	Level     OptionalNullID `json:"level"`
	Visible   *bool          `json:"visible"`
	SortOrder *int64         `json:"sort_order"`
}

// ---- business info ----

type BusinessInfo struct {
	LegalName      string  `json:"legal_name"`
	ICO            *string `json:"ico"`
	DIC            *string `json:"dic"`
	AddressLines   *string `json:"address_lines"`
	City           *string `json:"city"`
	Zip            *string `json:"zip"`
	Country        *string `json:"country"`
	RegisterNoteCS *string `json:"register_note_cs"`
	RegisterNoteEN *string `json:"register_note_en"`
	ContactEmail   *string `json:"contact_email"`
	ContactPhone   *string `json:"contact_phone"`
	DataNoteCS     *string `json:"data_note_cs"`
	DataNoteEN     *string `json:"data_note_en"`
	UpdatedAt      string  `json:"updated_at"`
}

type BusinessInfoUpdate struct {
	LegalName      *string `json:"legal_name"`
	ICO            *string `json:"ico"`
	DIC            *string `json:"dic"`
	AddressLines   *string `json:"address_lines"`
	City           *string `json:"city"`
	Zip            *string `json:"zip"`
	Country        *string `json:"country"`
	RegisterNoteCS *string `json:"register_note_cs"`
	RegisterNoteEN *string `json:"register_note_en"`
	ContactEmail   *string `json:"contact_email"`
	ContactPhone   *string `json:"contact_phone"`
	DataNoteCS     *string `json:"data_note_cs"`
	DataNoteEN     *string `json:"data_note_en"`
}

// ---- shared ----

// ReorderRequest is an ordered list of ids; the server rewrites sort_order to match.
type ReorderRequest struct {
	IDs []int64 `json:"ids"`
}

// AdminProjectsPage is the paginated admin project listing.
type AdminProjectsPage struct {
	Items      []ProjectDetail `json:"items"`
	NextCursor *string         `json:"next_cursor"`
}
