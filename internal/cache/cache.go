package cache

import (
	"fmt"
	"sync"

	"github.com/miekg/dns"
)


type Cache struct {
	entries map[string]string
	mu sync.RWMutex
	count int
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]string),
		count: 0,
	}
}

func (c *Cache) Add(response *dns.Msg) {
	domain := response.Question[0].Name
	var ips []string

    for _, ans := range response.Answer {
        switch record := ans.(type) {
        case *dns.A:
            ips = append(ips, record.A.String())
        // case *dns.AAAA:
        //    ips = append(ips, record.AAAA.String())
        }
    }

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.entries[domain]; !exists {
		c.entries[domain] = ips[0]
		c.count++
		fmt.Printf("Domain: %s, A record: %s is cached!", domain, ips[0])
	}
}

func (c *Cache) Get(domain string) (bool, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if _, exists := c.entries[domain]; !exists {
		return false, ""
	}

	ARecord := c.entries[domain]

	return true, ARecord
}

func (c *Cache) GetCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}