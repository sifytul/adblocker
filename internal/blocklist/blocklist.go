package blocklist

import (
	"strings"
	"sync"
)

type Blocklist struct {
	domains map[string]bool 
	mv sync.RWMutex
	count int
}

// NewBlocklist creates a new empty blocklist
func NewBlocklist() *Blocklist {
	return &Blocklist{
		domains: make(map[string]bool),
		count: 0,
	}
}

// normalization converts domain to standard format
func normalizeDomain(domain string) string {
	domain = strings.ToLower(domain)

	domain = strings.TrimSuffix(domain, ".")

	domain = strings.TrimSpace(domain)

	return domain
}

// Add adds a domain to the blocklist
func (b *Blocklist) Add(domain string) {
	domain = normalizeDomain(domain)

	b.mv.Lock()
	defer b.mv.Unlock()

	if !b.domains[domain] {
		b.domains[domain] = true
		b.count++
	}
}

// IsBlocked checks if a domain is in the blocklist
func (b *Blocklist) IsBlocked(domain string) bool {
	domain = normalizeDomain(domain)

	b.mv.RLock()
	defer b.mv.RUnlock()

	return b.domains[domain]
}

// Count returns th total number of blocked domains
func (b *Blocklist) Count() int {
	b.mv.RLock()
	defer b.mv.RUnlock()

	return b.count
}