// Package slug validates URL slugs against the fleet convention
// ^[a-z0-9][a-z0-9-]{0,N} (lowercase, digits, hyphens; must start alphanumeric).
package slug

import "regexp"

var (
	// Project slugs: up to 81 chars total (openapi Slug: {0,80}).
	projectRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,80}$`)
	// Category slugs: up to 41 chars total (PRD FR-4: {0,40}).
	categoryRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)
	// Game keys mirror the project pattern subset (openapi GameKey: {0,40}).
	gameKeyRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,40}$`)
)

// ValidProject reports whether s is a valid project slug.
func ValidProject(s string) bool { return projectRe.MatchString(s) }

// ValidCategory reports whether s is a valid category slug.
func ValidCategory(s string) bool { return categoryRe.MatchString(s) }

// ValidGameKey reports whether s is a valid arcade game key.
func ValidGameKey(s string) bool { return gameKeyRe.MatchString(s) }
