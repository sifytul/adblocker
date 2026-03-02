# DNS Ad Blocker - Project Documentation

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture Design](#architecture-design)
3. [Technical Specifications](#technical-specifications)
4. [Project Structure](#project-structure)
5. [Getting Started Checklist](#getting-started-checklist)
6. [Dependencies](#dependencies)
7. [Resources & References](#resources--references)

## 1. Project Overview

### 1.1 Project Name

**GoDNSAdBlocker** - A lightweight DNS-based ad blocker written in Go

### 1.2 Purpose

Build a local DNS resolver that blocks advertisements, trackers, and malicious domains by intercepting DNS queries and preventing resolution of blocked domains.

---

## 2. Architecture Design

### 2.1 High-Level Architecture

```
Client Device (Browser/App)
        |
        | DNS Query (e.g., ads.example.com)
        v
[GoDNSAdBlocker - DNS Server]
        |
        +---> [Query Handler]
        |         |
        |         v
        |   [Blocklist Checker]
        |         |
        |         +--> Blocked? --> Return 0.0.0.0
        |         |
        |         +--> Allowed? --> [Cache Checker]
        |                              |
        |                              v
        |                              +---> In Cache? --> Return cached Response
        |                              |
        |                              +---> Not in Cache? --> [Upstream Resolver]
        |                                                              |
        |                                                              v
        |                                                     External DNS (8.8.8.8)
        |                                                              |
        v                                                              v
   Response to Client  <--------------------------------------  Return actual IP
```

### 2.2 Core Components

1. **DNS Server**

   - Listens on UDP port 5333 (and optionally TCP)
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

5. **Cache Layer**

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
2. Not in blocklist? → Is domain in cache?
3. Not in cache? → Forward to upstream resolver
4. Error in query? → Return DNS error response
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

## 4. Project Structure

```
godnsblock/
├── cmd/
│   └── godnsadblocker/
│       └── main.go              # Application entry point
├── internal/
│   ├── server/
│   │   └── server.go            # DNS server implementation
│   ├── blocklist/
│   │   ├── blocklist.go         # Blocklist management
│   │   └── loader.go            # Blocklist loading logic
│   ├── logger/
│   │   └── logger.go            # Logging utilities
│   ├── resolver/
│   │   └── upstream.go          # Upstream DNS resolver
│   ├── cache/
│   │   └── cache.go             # DNS response cache
│   ├── config/
│   |   └── config.go            # Configuration management
│   └── stats/
│       └── stats.go             # Statistics tracking
├── configs/
│   └── config.yaml              # Default configuration
├── blocklists/
│   └── hosts.txt                # Local blocklist file
├── docs/
│   ├── installation.md
│   └── usage.md
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── Makefile
```

---

## 5. Getting Started Checklist

Before starting implementation:

- [ ] Read this entire document
- [ ] Set up Go development environment (Go 1.21+)
- [ ] Install `dig` or `nslookup` for testing
- [ ] Clone/create Git repository
- [ ] Read miekg/dns documentation
- [ ] Download sample blocklist file
- [ ] Understand your OS network settings (how to change DNS)

---

### Clone & Build

```bash
git clone https://github.com/sifytul/adblocker.git
cd adblocker
go mod tidy

# Build the application
go build -o build/godnsadblocker cmd/godnsadblocker/main.go
```

(Uses **miekg/dns**)

---

### 2️⃣ Run the Server

```bash
./build/godnsadblocker
```

---

## 🧪 Testing (Using Script)

Create a file:

```bash

echo "Testing normal domain..."
dig google.com @127.0.0.1 -p 5354 +short

echo "Testing blocked domain..."
dig doubleclick.net @127.0.0.1 -p 5354 +short

echo ""
echo "Test completed."
```

---

You should see:

- A real IP for `google.com`
- Empty / blocked response for `doubleclick.net`

---

## 6. Dependencies

### 6.1 Required Libraries

```go
github.com/miekg/dns        // DNS protocol handling
gopkg.in/yaml.v3            // YAML config parsing
```

---

## 7. Resources & References

### 7.1 DNS Protocol

- RFC 1035: Domain Names - Implementation and Specification
- https://www.ietf.org/rfc/rfc1035.txt

### 7.2 Go DNS Library

- miekg/dns documentation: https://pkg.go.dev/github.com/miekg/dns

### 7.3 Blocklists

- Steven Black's hosts: https://github.com/StevenBlack/hosts
- AdGuard DNS filter: https://github.com/AdguardTeam/AdGuardSDNSFilter
- OISD blocklist: https://oisd.nl/

### 7.4 Similar Projects (for inspiration)

- Pi-hole: https://github.com/pi-hole/pi-hole
- AdGuard Home: https://github.com/AdguardTeam/AdGuardHome
- Blocky: https://github.com/0xERR0R/blocky

---

## Notes

- This is a learning project - prioritize understanding over perfection
- Start simple, add complexity gradually
- Write tests as you go, not at the end
- Document decisions and challenges in a dev log
- Ask questions and research when stuck
- Have fun! This is a practical, useful project

---

**Last Updated:** March 2, 2026
**Author:** Md. Sifytul Karim
**Version:** 1.0
