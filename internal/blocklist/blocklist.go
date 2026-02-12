package blocklist

import (
	"bufio"
	"log"
	"os"
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

func (b *Blocklist) LoadFromFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return  err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()

		//Parse the line
		domain := parseHostsLine(line)
		if domain != "" {
			b.Add(domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d domains from %s", b.Count(), filepath)
	return nil
}

func parseHostsLine(line string) string {
	// Remove comments
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}

	// Trim whitespace
	line = strings.TrimSpace(line)

	// skip empty lines
	if line == "" {
		return ""
	}

	// Split by whitespaces: "0.0.0.0 ads.com" -> ["0.0.0.0", "ads.com"]
	parts := strings.Fields(line)

	// Need at least IP and domain
	if len(parts) < 2 {
		return ""
	}

	// Return the domain (second part)
	// parts[0] is the IP (0.0.0.0)
	// parts[1] is the domain we want
	return parts[1]
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