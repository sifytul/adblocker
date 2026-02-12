# DNS Ad Blocker - Project Documentation

## 1. Project Overview

### 1.1 Project Name
**GoDNSBlock** - A lightweight DNS-based ad blocker written in Go

### 1.2 Purpose
Build a local DNS resolver that blocks advertisements, trackers, and malicious domains by intercepting DNS queries and preventing resolution of blocked domains.

### 1.3 Goals
- Learn Go fundamentals (goroutines, channels, network programming)
- Understand DNS protocol and resolution process
- Build a practical, usable application
- Practice software architecture and design patterns

### 1.4 Target Users
- Developers learning Go
- Privacy-conscious users wanting network-wide ad blocking
- Users who want to understand how DNS-based blocking works

---

## 2. Architecture Design

### 2.1 High-Level Architecture

```
Client Device (Browser/App)
        |
        | DNS Query (e.g., ads.example.com)
        v
[GoDNSBlock - DNS Server]
        |
        +---> [Query Handler]
        |           |
        |           v
        |     [Blocklist Checker]
        |           |
        |           +---> Blocked? --> Return 0.0.0.0
        |           |
        |           +---> Allowed? --> [Upstream Resolver]
        |                                     |
        |                                     v
        |                              External DNS (8.8.8.8)
        |                                     |
        v                                     v
   Response to Client  <-------- Return actual IP
```

### 2.2 Core Components

1. **DNS Server**
   - Listens on UDP port 53 (and optionally TCP)
   - Receives DNS queries from clients
   - Manages concurrent connections using goroutines

2. **Query Handler**
   - Parses incoming DNS queries
   - Extracts domain name from query
   - Routes to blocklist checker

3. **Blocklist Manager**
   - Loads blocklists from files or URLs
   - Stores blocked domains efficiently (hash map)
   - Provides fast lookup capability
   - Supports periodic updates

4. **Upstream Resolver**
   - Forwards legitimate queries to external DNS servers
   - Handles DNS protocol communication
   - Returns responses to clients

5. **Cache Layer** (Optional - Phase 2)
   - Caches DNS responses to reduce upstream queries
   - Implements TTL-based expiration
   - Improves performance

6. **Configuration Manager**
   - Loads settings from config file
   - Manages upstream DNS servers
   - Controls blocklist sources
   - Sets logging levels

7. **Logger**
   - Records blocked queries
   - Logs errors and warnings
   - Provides statistics

---

## 3. Technical Specifications

### 3.1 DNS Protocol Basics

**DNS Query Flow:**
1. Client sends query: "What's the IP for example.com?"
2. DNS server looks up the domain
3. Server responds with IP address or error

**DNS Record Types We'll Handle:**
- A: IPv4 address
- AAAA: IPv6 address
- CNAME: Canonical name (alias)
- MX: Mail exchange
- TXT: Text records

### 3.2 Component Details

#### 3.2.1 DNS Server
```
Responsibilities:
- Bind to port 53 (requires root/admin privileges)
- Listen for UDP packets (DNS primarily uses UDP)
- Spawn goroutine for each incoming query
- Handle graceful shutdown

Key Functions:
- Start() - Initialize and start listening
- Stop() - Graceful shutdown
- HandleQuery(query) - Process individual queries
```

#### 3.2.2 Blocklist Manager
```
Responsibilities:
- Load blocklists from various sources
- Parse different formats (hosts file, domains list)
- Store in efficient data structure
- Provide O(1) lookup time

Data Sources:
- Local files
- Remote URLs (Steven Black's hosts, AdGuard filters)
- Custom user-defined lists

Storage Format:
- In-memory hash map: map[string]bool
- Key: domain name (normalized to lowercase)
- Value: true (blocked)
```

#### 3.2.3 Query Handler
```
Responsibilities:
- Parse DNS query packet
- Extract queried domain name
- Normalize domain (lowercase, trim)
- Check against blocklist
- Route to appropriate handler

Decision Logic:
1. Is domain in blocklist? → Return 0.0.0.0 or NXDOMAIN
2. Not in blocklist? → Forward to upstream resolver
3. Error in query? → Return DNS error response
```

#### 3.2.4 Upstream Resolver
```
Responsibilities:
- Maintain pool of upstream DNS servers
- Forward queries to external resolvers
- Handle timeouts and retries
- Return responses to clients

Default Upstream Servers:
- 8.8.8.8 (Google)
- 1.1.1.1 (Cloudflare)
- 9.9.9.9 (Quad9)

Features:
- Round-robin or failover selection
- Timeout handling (2-3 seconds)
- Retry logic for failed queries
```

---

## 4. Data Structures

### 4.1 Core Types

```go
// Config holds application configuration
type Config struct {
    ListenAddress    string   // e.g., "0.0.0.0:53"
    UpstreamServers  []string // e.g., ["8.8.8.8:53", "1.1.1.1:53"]
    BlocklistSources []string // URLs or file paths
    LogLevel         string   // "debug", "info", "warn", "error"
    CacheEnabled     bool
    CacheTTL         int      // seconds
}

// Blocklist manages blocked domains
type Blocklist struct {
    domains map[string]bool
    mu      sync.RWMutex  // For concurrent access
    count   int           // Total blocked domains
}

// DNSServer represents the main server
type DNSServer struct {
    config    *Config
    blocklist *Blocklist
    resolver  *UpstreamResolver
    server    *dns.Server
}

// UpstreamResolver handles forwarding queries
type UpstreamResolver struct {
    servers []string
    client  *dns.Client
    index   int  // For round-robin
}

// Statistics for monitoring
type Stats struct {
    TotalQueries   int64
    BlockedQueries int64
    AllowedQueries int64
    CacheHits      int64
    StartTime      time.Time
}
```

### 4.2 Key Interfaces

```go
// BlocklistProvider defines how to load blocklists
type BlocklistProvider interface {
    Load() ([]string, error)
    Source() string
}

// Resolver defines DNS resolution behavior
type Resolver interface {
    Resolve(query *dns.Msg) (*dns.Msg, error)
}
```

---

## 5. Implementation Plan

### Phase 1: Basic DNS Server (Week 1)
**Goal:** Get a minimal DNS server running that forwards all queries

Tasks:
1. Set up Go project structure
2. Install `github.com/miekg/dns` library
3. Create basic DNS server that listens on port 53
4. Implement simple query forwarding to 8.8.8.8
5. Test with `dig` or `nslookup`

**Deliverable:** A DNS server that successfully resolves any domain

---

### Phase 2: Blocklist Integration (Week 2)
**Goal:** Add blocking functionality

Tasks:
1. Create Blocklist struct with map storage
2. Implement file-based blocklist loading (hosts format)
3. Add domain lookup logic in query handler
4. Return 0.0.0.0 for blocked domains
5. Download and test with Steven Black's hosts file

**Deliverable:** DNS server that blocks ads based on blocklist

---

### Phase 3: Configuration & Logging (Week 3)
**Goal:** Make the application configurable and observable

Tasks:
1. Create configuration file (YAML or JSON)
2. Implement config loading
3. Add structured logging (use `log/slog` or `zap`)
4. Log blocked queries and statistics
5. Add command-line flags

**Deliverable:** Configurable DNS blocker with logging

---

### Phase 4: Advanced Features (Week 4)
**Goal:** Add polish and performance improvements

Tasks:
1. Implement basic caching
2. Add multiple upstream resolver support
3. Support remote blocklist URLs with auto-update
4. Add statistics endpoint (HTTP server)
5. Implement graceful shutdown
6. Add systemd service file for Linux

**Deliverable:** Production-ready DNS ad blocker

---

### Phase 5: Testing & Documentation (Week 5)
**Goal:** Ensure reliability and usability

Tasks:
1. Write unit tests for core components
2. Add integration tests
3. Create user documentation (README)
4. Add installation guide
5. Performance testing and optimization

**Deliverable:** Well-tested and documented project

---

## 6. Testing Strategy

### 6.1 Unit Tests
- Blocklist operations (add, check, load)
- Domain normalization
- Config loading
- Query parsing

### 6.2 Integration Tests
- Full DNS query flow
- Upstream resolver failover
- Blocklist updates
- Cache functionality

### 6.3 Manual Testing Tools
- `dig @localhost example.com` - Query the DNS server
- `nslookup ads.google.com localhost` - Test blocking
- Wireshark - Inspect DNS packets
- Browser - Real-world testing

### 6.4 Performance Testing
- Load testing with multiple concurrent queries
- Memory usage monitoring
- Response time measurements
- Cache hit rate analysis

---

## 7. Security Considerations

### 7.1 Potential Risks
1. **DNS Spoofing** - Attacker could poison cache
2. **Amplification Attacks** - DNS server used in DDoS
3. **Privacy Leaks** - Logging sensitive query data
4. **Privilege Escalation** - Running on port 53 requires root

### 7.2 Mitigations
1. Validate DNS responses from upstream
2. Rate limiting per client IP
3. Minimal logging, anonymize IPs if needed
4. Drop privileges after binding to port 53
5. Input validation on all queries

---

## 8. Future Enhancements

### 8.1 Short-term
- Web UI for statistics and management
- Whitelist support (override blocklist)
- Custom block page (for HTTP requests)
- Regex-based blocking rules
- Category-based filtering

### 8.2 Long-term
- DNS-over-HTTPS (DoH) support
- DNS-over-TLS (DoT) support
- Distributed blocklist sharing
- Machine learning for automatic threat detection
- DNSSEC validation
- IPv6 full support
- Docker containerization
- Metrics export (Prometheus)

---

## 9. Project Structure

```
godnsblock/
├── cmd/
│   └── godnsblock/
│       └── main.go              # Application entry point
├── internal/
│   ├── server/
│   │   └── server.go            # DNS server implementation
│   ├── blocklist/
│   │   ├── blocklist.go         # Blocklist management
│   │   └── loader.go            # Blocklist loading logic
│   ├── resolver/
│   │   └── upstream.go          # Upstream DNS resolver
│   ├── cache/
│   │   └── cache.go             # DNS response cache
│   └── config/
│       └── config.go            # Configuration management
├── pkg/
│   └── stats/
│       └── stats.go             # Statistics tracking
├── configs/
│   └── config.yaml              # Default configuration
├── blocklists/
│   └── hosts.txt                # Local blocklist file
├── docs/
│   ├── installation.md
│   └── usage.md
├── tests/
│   ├── integration/
│   └── unit/
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── Makefile
```

---

## 10. Dependencies

### 10.1 Required Libraries
```go
github.com/miekg/dns        // DNS protocol handling
gopkg.in/yaml.v3            // YAML config parsing
```

### 10.2 Optional Libraries
```go
github.com/sirupsen/logrus  // Advanced logging
github.com/spf13/cobra      // CLI framework
github.com/prometheus/client_golang // Metrics
```

---

## 11. Resources & References

### 11.1 DNS Protocol
- RFC 1035: Domain Names - Implementation and Specification
- https://www.ietf.org/rfc/rfc1035.txt

### 11.2 Go DNS Library
- miekg/dns documentation: https://pkg.go.dev/github.com/miekg/dns

### 11.3 Blocklists
- Steven Black's hosts: https://github.com/StevenBlack/hosts
- AdGuard DNS filter: https://github.com/AdguardTeam/AdGuardSDNSFilter
- OISD blocklist: https://oisd.nl/

### 11.4 Similar Projects (for inspiration)
- Pi-hole: https://github.com/pi-hole/pi-hole
- AdGuard Home: https://github.com/AdguardTeam/AdGuardHome
- Blocky: https://github.com/0xERR0R/blocky

---

## 12. Success Metrics

### 12.1 Functional Goals
- [ ] Successfully blocks ads on websites
- [ ] Handles 100+ concurrent queries
- [ ] Response time under 50ms
- [ ] Blocks at least 100,000 domains
- [ ] Zero crashes in 24-hour test

### 12.2 Learning Goals
- [ ] Understand DNS protocol deeply
- [ ] Master goroutines and channels
- [ ] Learn network programming in Go
- [ ] Practice concurrent data structures
- [ ] Build production-ready Go application

---

## 13. Getting Started Checklist

Before starting implementation:
- [ ] Read this entire document
- [ ] Set up Go development environment (Go 1.21+)
- [ ] Install `dig` or `nslookup` for testing
- [ ] Clone/create Git repository
- [ ] Read miekg/dns documentation
- [ ] Download sample blocklist file
- [ ] Understand your OS network settings (how to change DNS)

---

## Notes

- This is a learning project - prioritize understanding over perfection
- Start simple, add complexity gradually
- Write tests as you go, not at the end
- Document decisions and challenges in a dev log
- Ask questions and research when stuck
- Have fun! This is a practical, useful project

---

**Last Updated:** February 10, 2026
**Author:** Md. Sifytul Karim
**Version:** 1.0
