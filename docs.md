# GoDNSBlock - DNS-Based Ad Blocker

A lightweight, efficient DNS-based ad blocker written in Go. Block ads, trackers, and malicious domains at the network level by intercepting DNS queries.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-in%20development-yellow)](https://github.com/yourusername/godnsblock)

---

## 🚀 Features

✅ **Implemented (Phase 1 & 2)**

- DNS server with UDP protocol support
- Domain blocklist management with O(1) lookup
- Upstream DNS forwarding (Google DNS, Cloudflare, etc.)
- Case-insensitive domain matching
- Hosts file format support
- Thread-safe concurrent query handling
- Returns `0.0.0.0` for blocked domains

🚧 **In Progress**

- YAML configuration file support
- Command-line flags
- Structured logging with levels
- Query statistics tracking
- Multiple blocklist sources

📋 **Planned**

- DNS response caching
- Web UI for management
- Remote blocklist auto-updates
- Whitelist support
- DNS-over-HTTPS (DoH)
- Docker support

---

## 📖 Table of Contents

- [How It Works](#how-it-works)
- [Architecture](#architecture)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Project Structure](#project-structure)
- [Development](#development)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

---

## 🔍 How It Works

GoDNSBlock acts as a local DNS resolver that sits between your device and the internet:

```
┌─────────────┐
│   Browser   │
│   or App    │
└──────┬──────┘
       │ 1. DNS Query: "ads.example.com"
       ▼
┌─────────────────┐
│  GoDNSBlock     │
│  DNS Server     │
└────────┬────────┘
         │ 2. Check Blocklist
         ▼
    ┌────────┐
    │Blocked?│
    └───┬─┬──┘
        │ │
   YES  │ │  NO
        │ │
        ▼ ▼
┌───────────┐  ┌──────────────┐
│Return     │  │Forward to    │
│0.0.0.0    │  │Upstream DNS  │
└───────────┘  │(8.8.8.8)     │
               └──────┬───────┘
                      │ 3. Get Real IP
                      ▼
               ┌──────────────┐
               │Return Real IP│
               │140.82.121.4  │
               └──────────────┘
```

### The Process

1. **Device makes DNS query**: When you visit a website, your device asks "What's the IP for this domain?"

2. **GoDNSBlock intercepts**: Instead of going directly to the internet, the query comes to our DNS server

3. **Blocklist check**: We check if the domain is in our blocklist (ads, trackers, malware)

   - **If blocked**: Return `0.0.0.0` → Connection fails → Ad doesn't load! ✅
   - **If allowed**: Forward to upstream DNS → Get real IP → Website loads normally

4. **Fast & efficient**: All checks happen in memory with O(1) lookup time

---

## 🏗️ Architecture

### High-Level Components

```
┌─────────────────────────────────────────────────────────┐
│                     GoDNSBlock                          │
│                                                         │
│  ┌──────────────┐         ┌──────────────┐            │
│  │  DNS Server  │────────▶│Query Handler │            │
│  │  (Port 53)   │         │              │            │
│  └──────────────┘         └──────┬───────┘            │
│                                   │                     │
│                                   ▼                     │
│                          ┌────────────────┐            │
│                          │Blocklist Check │            │
│                          │  (Hash Map)    │            │
│                          └────────┬───────┘            │
│                                   │                     │
│                        ┌──────────┴──────────┐         │
│                        │                     │         │
│                   Blocked?              Allowed?       │
│                        │                     │         │
│                        ▼                     ▼         │
│              ┌─────────────────┐   ┌────────────────┐ │
│              │ Return 0.0.0.0  │   │Upstream        │ │
│              │                 │   │Resolver        │ │
│              └─────────────────┘   │(Forward Query) │ │
│                                    └────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Core Components

#### 1. **DNS Server** (`internal/server/`)

- Binds to UDP port 53
- Listens for DNS queries
- Spawns goroutine for each query (concurrent handling)
- Routes queries to handler

#### 2. **Query Handler** (`internal/server/`)

- Parses DNS query packets
- Extracts domain name
- Checks against blocklist
- Creates blocked response or forwards query

#### 3. **Blocklist Manager** (`internal/blocklist/`)

- Loads domains from hosts files
- Stores in hash map: `map[string]bool`
- Provides O(1) lookup: `IsBlocked(domain)`
- Thread-safe with `sync.RWMutex`

#### 4. **Upstream Resolver** (`internal/resolver/`)

- Forwards allowed queries to real DNS (8.8.8.8)
- Handles DNS protocol communication
- Returns responses to clients
- Implements timeout (3 seconds)

---

## 📦 Installation

### Prerequisites

- **Go 1.21+** - [Download here](https://go.dev/dl/)
- **Root/Admin privileges** - Required to bind to port 53
- **Linux/Mac/Windows** - Cross-platform support

### Quick Start

```bash
# Clone the repository
git clone https://github.com/yourusername/godnsblock.git
cd godnsblock

# Install dependencies
go mod download

# Build the project
go build -o godnsblock cmd/godnsblock/main.go

# Run (requires sudo/admin)
sudo ./godnsblock
```

### Alternative: Run without building

```bash
sudo go run cmd/godnsblock/main.go
```

---

## 🎯 Usage

### Starting the Server

```bash
# Start with default settings
sudo ./godnsblock

# Expected output:
# Loading blocklist...
# Loaded 13 domains from blocklists/test-blocklist.txt
# Starting DNS server on 0.0.0.0:53
# Server started successfully
# DNS server is running. Press Ctrl+C to stop.
```

### Testing with Command Line

#### Test blocked domain:

```bash
dig @localhost ads.google.com

# Expected:
# ;; ANSWER SECTION:
# ads.google.com.    300    IN    A    0.0.0.0
```

#### Test allowed domain:

```bash
dig @localhost github.com

# Expected:
# ;; ANSWER SECTION:
# github.com.    60    IN    A    140.82.121.4
```

#### Using nslookup (Windows):

```cmd
nslookup ads.google.com 127.0.0.1
```

### System-Wide Ad Blocking

To use GoDNSBlock for all applications on your device:

#### Linux/Mac:

```bash
# Edit network settings to use 127.0.0.1 as DNS server
# Or edit /etc/resolv.conf:
echo "nameserver 127.0.0.1" | sudo tee /etc/resolv.conf
```

#### Windows:

1. Open Network Connections
2. Right-click your connection → Properties
3. Select "Internet Protocol Version 4 (TCP/IPv4)"
4. Set DNS server to: `127.0.0.1`

#### Note:

Remember to start GoDNSBlock before changing DNS settings!

---

## ⚙️ Configuration

### Current Configuration (Hardcoded)

Currently in `cmd/godnsblock/main.go`:

```go
listenAddr := "0.0.0.0:53"           // Listen on all interfaces
upstreamDNS := "8.8.8.8:53"          // Google DNS
blocklistFile := "blocklists/test-blocklist.txt"
```

### Planned: YAML Configuration (Phase 3)

```yaml
# config.yaml (coming soon)
server:
  listen_address: "0.0.0.0:53"

upstream:
  servers:
    - "8.8.8.8:53" # Google DNS
    - "1.1.1.1:53" # Cloudflare DNS
    - "9.9.9.9:53" # Quad9 DNS

blocklists:
  sources:
    - path: "blocklists/ads.txt"
    - path: "blocklists/trackers.txt"
    - url: "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"

logging:
  level: "info" # debug, info, warn, error
  log_queries: true
  log_blocked: true
  output: "logs/godnsblock.log"
```

---

## 📁 Project Structure

```
godnsblock/
├── cmd/
│   └── godnsblock/
│       └── main.go              # Application entry point
│
├── internal/
│   ├── server/
│   │   └── server.go            # DNS server implementation
│   │
│   ├── blocklist/
│   │   ├── blocklist.go         # Blocklist data structure
│   │   └── loader.go            # Hosts file parser
│   │
│   └── resolver/
│       └── upstream.go          # Upstream DNS resolver
│
├── blocklists/
│   └── test-blocklist.txt       # Sample blocklist
│
├── go.mod                        # Go module dependencies
├── go.sum                        # Dependency checksums
├── README.md                     # This file
└── LICENSE                       # Project license
```

### Key Files Explained

**`cmd/godnsblock/main.go`**

- Entry point of the application
- Initializes server and blocklist
- Handles graceful shutdown

**`internal/server/server.go`**

- DNS server logic
- Query handling
- Blocking/forwarding decisions

**`internal/blocklist/blocklist.go`**

- Blocklist data structure
- Thread-safe operations
- Domain checking logic

**`internal/blocklist/loader.go`**

- Parses hosts file format
- Loads domains into memory
- Handles normalization

**`internal/resolver/upstream.go`**

- Forwards queries to external DNS
- Manages DNS client
- Handles timeouts

---

## 🛠️ Development

### Setting Up Development Environment

```bash
# Clone the repository
git clone https://github.com/yourusername/godnsblock.git
cd godnsblock

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Run with race detection
go run -race cmd/godnsblock/main.go

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o godnsblock-linux cmd/godnsblock/main.go
GOOS=darwin GOARCH=amd64 go build -o godnsblock-mac cmd/godnsblock/main.go
GOOS=windows GOARCH=amd64 go build -o godnsblock.exe cmd/godnsblock/main.go
```

### Key Go Concepts Used

**Goroutines & Concurrency**

```go
// Each DNS query runs in its own goroutine
go func() {
    s.server.ListenAndServe()
}()
```

**Channels**

```go
// Used for graceful shutdown
sig := make(chan os.Signal, 1)
signal.Notify(sig, syscall.SIGINT)
```

**Mutexes for Thread Safety**

```go
// Blocklist is accessed by multiple goroutines
b.mu.RLock()
defer b.mu.RUnlock()
return b.domains[domain]
```

**Maps for Fast Lookup**

```go
// O(1) domain checking
domains map[string]bool
```

---

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/blocklist/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Manual Testing

#### Test 1: Verify server is running

```bash
sudo ./godnsblock
# Should see: "DNS server is running"
```

#### Test 2: Check blocked domain

```bash
dig @localhost ads.google.com +short
# Expected: 0.0.0.0
```

#### Test 3: Check allowed domain

```bash
dig @localhost google.com +short
# Expected: Real IP addresses (e.g., 142.250.80.46)
```

#### Test 4: Case insensitivity

```bash
dig @localhost ADS.GOOGLE.COM +short
dig @localhost Ads.Google.Com +short
# Both should return: 0.0.0.0
```

#### Test 5: Performance test

```bash
# Install dnsperf (DNS benchmarking tool)
# Then test queries per second
dnsperf -s localhost -d queryfile.txt
```

### Monitoring Logs

Watch logs in real-time:

```bash
sudo ./godnsblock | grep BLOCKED
# Shows only blocked queries

sudo ./godnsblock | grep ALLOWED
# Shows only allowed queries
```

---

## 🎓 Learning Resources

This project is great for learning Go! Here are the key concepts:

### Go Concepts Demonstrated

1. **Network Programming**

   - UDP socket handling
   - DNS protocol implementation
   - Concurrent connection handling

2. **Concurrency**

   - Goroutines for parallel query processing
   - Channels for communication
   - Mutexes for thread-safe data structures

3. **Data Structures**

   - Hash maps for O(1) lookups
   - Structs for organizing data
   - Interfaces for abstraction

4. **File I/O**

   - Reading blocklist files
   - Parsing text formats
   - Error handling

5. **Package Organization**
   - Internal packages
   - Separation of concerns
   - Clean architecture

### Recommended Reading

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Concurrency Patterns](https://go.dev/blog/pipelines)
- [DNS RFC 1035](https://www.ietf.org/rfc/rfc1035.txt)
- [miekg/dns Documentation](https://pkg.go.dev/github.com/miekg/dns)

---

## 📊 Statistics & Performance

### Current Capabilities

- **Blocklist Size**: 13 domains (test file) → 100,000+ (production ready)
- **Lookup Speed**: O(1) - instant
- **Memory Usage**: ~1MB per 100,000 domains
- **Query Handling**: Concurrent (goroutines)
- **Response Time**: <10ms (local) + upstream latency

### Benchmarks (Coming Soon)

```
BenchmarkBlocklistLookup     1000000    1.2 ns/op
BenchmarkQueryHandling       10000      120 µs/op
BenchmarkConcurrentQueries   5000       240 µs/op
```

---

## 🔒 Security Considerations

### Current Implementation

✅ **Implemented**

- Input validation on DNS queries
- Thread-safe data structures
- Proper error handling

⚠️ **TODO**

- Drop root privileges after binding port 53
- Rate limiting per client IP
- DNSSEC validation
- DNS-over-TLS support

### Best Practices

1. **Run with minimal privileges**: Drop root after binding port 53
2. **Keep blocklists updated**: Regularly refresh from trusted sources
3. **Monitor logs**: Watch for unusual query patterns
4. **Whitelist critical domains**: Don't break important services

---

## 🚀 Roadmap

### ✅ Phase 1: Basic DNS Server (Completed)

- UDP DNS server
- Query forwarding
- Upstream resolver

### ✅ Phase 2: Blocklist Integration (Completed)

- Hosts file parsing
- Domain blocking
- Case-insensitive matching

### 🚧 Phase 3: Configuration & Logging (In Progress)

- YAML configuration
- Command-line flags
- Structured logging
- Statistics tracking

### 📋 Phase 4: Advanced Features (Planned)

- DNS response caching
- Multiple upstream resolvers
- Remote blocklist updates
- Whitelist support

### 📋 Phase 5: Production Ready (Planned)

- Web UI dashboard
- Prometheus metrics
- Docker support
- Systemd service file
- Auto-update mechanism

### 📋 Phase 6: Advanced Security (Future)

- DNS-over-HTTPS (DoH)
- DNS-over-TLS (DoT)
- DNSSEC validation
- Rate limiting
- Query logging with privacy

---

## 🤝 Contributing

Contributions are welcome! This is a learning project, so don't worry about perfection.

### How to Contribute

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Commit your changes**: `git commit -m 'Add amazing feature'`
4. **Push to the branch**: `git push origin feature/amazing-feature`
5. **Open a Pull Request**

### Development Guidelines

- Write tests for new features
- Follow Go conventions (run `gofmt`)
- Update documentation
- Add comments for complex logic
- Keep functions small and focused

### Areas That Need Help

- [ ] Web UI implementation
- [ ] More blocklist sources
- [ ] Performance optimizations
- [ ] Better error messages
- [ ] Documentation improvements
- [ ] Example configurations
- [ ] Docker containerization

---

## 📝 Blocklist Sources

### Current

- `blocklists/test-blocklist.txt` - Small test file (13 domains)

### Recommended Production Sources

**Steven Black's Hosts** (100,000+ domains)

```
https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts
```

**AdGuard DNS Filter** (Popular choice)

```
https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt
```

**OISD Blocklist** (Aggressive blocking)

```
https://raw.githubusercontent.com/sjhgvr/oisd/main/domainswild.txt
```

**Peter Lowe's List** (Ad servers only)

```
https://pgl.yoyo.org/adservers/serverlist.php?hostformat=hosts&showintro=0
```

### How to Add More Blocklists

1. Download the file
2. Place in `blocklists/` directory
3. Update `blocklistFile` in `main.go`
4. Restart server

---

## 🐛 Troubleshooting

### Common Issues

#### Port 53 Already in Use

**Problem**: `bind: address already in use`

**Solution**:

```bash
# Linux: Check what's using port 53
sudo lsof -i :53

# Stop systemd-resolved
sudo systemctl stop systemd-resolved

# Or run on different port (testing only)
# Modify main.go: listenAddr := "0.0.0.0:5353"
```

#### Permission Denied

**Problem**: `bind: permission denied`

**Solution**: Run with sudo/admin privileges

```bash
sudo ./godnsblock
```

#### No Logs Appearing

**Problem**: Queries work but no logs

**Possible causes**:

1. Another DNS service is handling queries
2. Querying wrong IP (use `dig @127.0.0.1` or your actual IP)
3. systemd-resolved intercepting on 127.0.0.53

**Solution**: Test with actual IP address

```bash
# Find your IP
ip addr show
# Then query it directly
dig @192.168.1.100 google.com
```

#### Timeouts

**Problem**: DNS queries timeout

**Solution**:

- Check upstream DNS is reachable: `ping 8.8.8.8`
- Verify firewall allows UDP port 53
- Check internet connection

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **[miekg/dns](https://github.com/miekg/dns)** - Excellent Go DNS library
- **[Pi-hole](https://pi-hole.net/)** - Inspiration for DNS-based blocking
- **[AdGuard Home](https://adguard.com/en/adguard-home/overview.html)** - Feature ideas
- **[Steven Black](https://github.com/StevenBlack/hosts)** - Comprehensive blocklists

---

## 📞 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/godnsblock/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/godnsblock/discussions)
- **Email**: your.email@example.com

---

## 📈 Project Status

**Current Version**: 0.2.0 (Phase 2 Complete)

**Development Status**: 🟡 Active Development

**What Works**:

- ✅ DNS query interception
- ✅ Domain blocking with blocklists
- ✅ Upstream forwarding
- ✅ Concurrent query handling
- ✅ Thread-safe operations

**In Progress**:

- 🚧 Configuration system
- 🚧 Advanced logging
- 🚧 Statistics tracking

**Coming Soon**:

- 📋 DNS caching
- 📋 Web UI
- 📋 Remote blocklists
- 📋 Docker support

---

## 🎯 Quick Reference

### Commands Cheat Sheet

```bash
# Build
go build -o godnsblock cmd/godnsblock/main.go

# Run
sudo ./godnsblock

# Test blocked domain
dig @localhost ads.google.com

# Test allowed domain
dig @localhost google.com

# Run tests
go test ./...

# View logs in real-time
sudo ./godnsblock | grep BLOCKED

# Stop server
Ctrl+C
```

### Important Files

| File                              | Purpose           |
| --------------------------------- | ----------------- |
| `cmd/godnsblock/main.go`          | Entry point       |
| `internal/server/server.go`       | DNS server logic  |
| `internal/blocklist/blocklist.go` | Blocklist manager |
| `internal/resolver/upstream.go`   | Upstream resolver |
| `blocklists/test-blocklist.txt`   | Test domains      |

---

**Made with ❤️ and Go**

_Learning project for understanding DNS, network programming, and Go concurrency_

---

**Last Updated**: February 13, 2026
