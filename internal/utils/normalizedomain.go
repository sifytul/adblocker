package utils

import "strings"

// normalization converts domain to standard format
func NormalizeDomain(domain string) string {
	domain = strings.ToLower(domain)

	domain = strings.TrimSuffix(domain, ".")

	domain = strings.TrimSpace(domain)

	return domain
}