# GoDNSBlock - Phase 3 Documentation Update

## Phase 3: Configuration & Logging - COMPLETED ✅

### Overview

Phase 3 transformed the DNS ad blocker from a hardcoded prototype into a flexible, production-ready application with comprehensive configuration management and observability features.

---

## What Was Added in Phase 3

### 1. **Configuration System**

- YAML-based configuration files
- Structured config with validation
- Default configuration fallbacks
- Easy customization without recompiling

### 2. **Command-Line Interface**

- CLI flags for runtime overrides
- Version information
- Flexible configuration loading

### 3. **Structured Logging**

- Log levels (DEBUG, INFO, WARN, ERROR)
- Timestamped log entries
- Configurable log output (stdout or file)
- Separate logging for blocked vs allowed queries

### 4. **Statistics Tracking**

- Real-time query counting
- Block rate calculation
- Uptime tracking
- Queries per second (QPS) metrics
- Periodic statistics reporting

---

## New Project Structure

```
godnsblock/
├── cmd/
│   └── godnsblock/
│       └── main.go              # ✅ Updated: CLI flags, config loading
│
├── internal/
│   ├── server/
│   │   └── server.go            # ✅ Updated: Uses config, logger, stats
│   │
│   ├── blocklists/
│   │   ├── blocklist.go         # Same as Phase 2
│   │   └── loader.go            # Same as Phase 2
│   │
│   ├── resolver/
│   │   └── upstream.go          # Same as Phase 2
│   │
│   ├── config/                  # 🆕 NEW in Phase 3
│   │   └── config.go            # Configuration structure and loading
│   │
│   ├── logger/                  # 🆕 NEW in Phase 3
│   │   └── logger.go            # Structured logging implementation
│   │
│   └── stats/                   # 🆕 NEW in Phase 3
│       └── stats.go             # Statistics tracking
│
├── configs/                     # 🆕 NEW in Phase 3
│   └── config.yaml              # Default configuration file
│
├── blocklists/
│   └── test-blocklist.txt
│
├── logs/                        # 🆕 NEW in Phase 3 (optional)
│   └── godnsblock.log           # Log output file (if configured)
│
├── go.mod
├── go.sum
└── README.md
```

---

## Configuration File Reference

### Complete `config.yaml` Structure

```yaml
# GoDNSBlock Configuration File

server:
  # Address to listen on
  # Examples:
  #   "0.0.0.0:53"    - Listen on all interfaces
  #   "127.0.0.1:53"  - Listen only on localhost
  #   "0.0.0.0:5353"  - Non-standard port (no root required)
  listen_address: "0.0.0.0:53"

  # Protocol: "udp" or "tcp"
  # UDP is standard for DNS, TCP for large responses
  protocol: "udp"

upstream:
  # List of upstream DNS servers
  # Queries for allowed domains are forwarded here
  servers:
    - "8.8.8.8:53" # Google Public DNS
    - "1.1.1.1:53" # Cloudflare DNS
    - "9.9.9.9:53" # Quad9 DNS

  # Timeout for upstream queries (seconds)
  # Recommended: 2-5 seconds
  timeout: 3

blocklist:
  # Sources for blocked domains
  # Can be local file paths
  # (URLs will be supported in Phase 4)
  sources:
    - "blocklists/test-blocklist.txt"
    - "blocklists/ads.txt"
    - "blocklists/trackers.txt"
    - "blocklists/adult-blocking.txt"

logging:
  # Log level: "debug", "info", "warn", "error"
  # debug - All messages (very verbose)
  # info  - General information
  # warn  - Warnings only
  # error - Errors only
  level: "info"

  # Log all DNS queries (including allowed ones)
  # Set to false in production to reduce log volume
  log_queries: true

  # Log blocked queries
  # Recommended: true (see what's being blocked)
  log_blocked: true

  # Output file path
  # Empty string ("") = stdout
  # Or specify file: "logs/godnsblock.log"
  output_file: ""
```

### Configuration Validation Rules

The system validates configuration on load:

**Server Section:**

- `listen_address` cannot be empty
- `protocol` must be "udp" or "tcp"

**Upstream Section:**

- At least one server required
- `timeout` must be positive (> 0)

**Blocklist Section:**

- At least one source required

**Logging Section:**

- `level` must be one of: debug, info, warn, error

**Invalid configurations will fail with clear error messages.**

---

## Command-Line Interface

### Available Flags

```bash
--config string      Path to configuration file (default "configs/config.yaml")
--listen string      Override listen address (e.g., "0.0.0.0:5353")
--log-level string   Override log level (debug, info, warn, error)
--version           Show version and exit
```

### Usage Examples

**1. Run with default configuration:**

```bash
sudo ./godnsblock
```

**2. Use custom config file:**

```bash
sudo ./godnsblock --config /etc/godnsblock/prod.yaml
```

**3. Override listen address (useful for testing without root):**

```bash
./godnsblock --listen 0.0.0.0:5353
```

**4. Enable debug logging temporarily:**

```bash
sudo ./godnsblock --log-level debug
```

**5. Combine multiple flags:**

```bash
sudo ./godnsblock --config prod.yaml --listen 0.0.0.0:53 --log-level warn
```

**6. Show version:**

```bash
./godnsblock --version
# Output: GoDNSBlock v0.3.0
```

---

## Logging System

### Log Format

All log entries follow this format:

```
[TIMESTAMP] LEVEL: MESSAGE
```

Example:

```
[2024-02-13 15:04:05] INFO: DNS server is running
[2024-02-13 15:04:06] INFO: BLOCKED: ads.google.com. (from 192.168.1.100:54321)
[2024-02-13 15:04:07] DEBUG: ALLOWED: github.com. (from 192.168.1.100:54322)
[2024-02-13 15:04:08] ERROR: Failed to resolve domain: timeout
```

### Log Levels

**DEBUG** - Verbose details for troubleshooting

- All queries (blocked and allowed)
- Internal operations
- Use for: Development, debugging issues

**INFO** - General operational information

- Server start/stop
- Blocklist loading
- Blocked queries
- Statistics
- Use for: Normal operation, monitoring

**WARN** - Warning conditions

- Configuration issues
- Non-critical errors
- Use for: Alerting to potential problems

**ERROR** - Error conditions

- Failed DNS resolutions
- Network errors
- Critical failures
- Use for: Monitoring, alerting

### Log Output Options

**Option 1: Standard Output (Terminal)**

```yaml
logging:
  output_file: ""
```

Logs printed to console. Good for:

- Development
- Docker containers
- Systemd with journald

**Option 2: File Output**

```yaml
logging:
  output_file: "logs/godnsblock.log"
```

Logs written to file. Good for:

- Production servers
- Log rotation
- Long-term storage

**Create log directory:**

```bash
mkdir -p logs
```

---

## Statistics System

### Tracked Metrics

**1. Uptime**

- How long the server has been running
- Format: Hours, minutes, seconds

**2. Total Queries**

- All DNS queries received

**3. Blocked Queries**

- Queries blocked by blocklist
- Percentage of total

**4. Allowed Queries**

- Queries forwarded to upstream

**5. Errors**

- Failed resolutions
- Network errors

**6. Queries Per Second (QPS)**

- Average query rate

### Statistics Display

Statistics are automatically logged every 60 seconds:

```
[2024-02-13 15:05:05] INFO: Statistics: Uptime: 1m0s | Total: 150 | Blocked: 45 (30.0%) | Allowed: 105 | Errors: 0 | QPS: 2.50
```

**Interpreting the stats:**

- **Uptime: 1m0s** - Server running for 1 minute
- **Total: 150** - 150 queries received
- **Blocked: 45 (30.0%)** - 45 queries blocked (30% block rate)
- **Allowed: 105** - 105 queries forwarded
- **Errors: 0** - No errors
- **QPS: 2.50** - Averaging 2.5 queries per second

### Final Statistics

When shutting down (Ctrl+C), final statistics are displayed:

```
^C[2024-02-13 16:00:00] INFO: Shutting down...
[2024-02-13 16:00:00] INFO: Final statistics: Uptime: 55m55s | Total: 8325 | Blocked: 2498 (30.0%) | Allowed: 5827 | Errors: 0 | QPS: 2.48
[2024-02-13 16:00:00] INFO: Shutdown complete
```

---

## Sample Configuration Scenarios

### Development Configuration

**Purpose:** Local testing, verbose logging, non-privileged port

```yaml
server:
  listen_address: "127.0.0.1:5353" # Localhost only, non-root port
  protocol: "udp"

upstream:
  servers: ["8.8.8.8:53"]
  timeout: 3

blocklist:
  sources: ["blocklists/test-blocklist.txt"]

logging:
  level: "debug" # See everything
  log_queries: true # Log all queries
  log_blocked: true
  output_file: "" # Console output
```

**Run without sudo:**

```bash
./godnsblock --config dev-config.yaml
```

**Test:**

```bash
dig @localhost -p 5353 ads.google.com
```

---

### Production Configuration

**Purpose:** Production server, minimal logging, multiple upstreams

```yaml
server:
  listen_address: "0.0.0.0:53"
  protocol: "udp"

upstream:
  servers:
    - "8.8.8.8:53"
    - "1.1.1.1:53"
    - "9.9.9.9:53"
  timeout: 3

blocklist:
  sources:
    - "blocklists/ads.txt"
    - "blocklists/trackers.txt"
    - "blocklists/malware.txt"

logging:
  level: "warn" # Only warnings and errors
  log_queries: false # Don't log every query
  log_blocked: false # Don't log blocks (privacy)
  output_file: "/var/log/godnsblock/godnsblock.log"
```

**Setup:**

```bash
sudo mkdir -p /var/log/godnsblock
sudo ./godnsblock --config /etc/godnsblock/config.yaml
```

---

### Privacy-Focused Configuration

**Purpose:** Minimal logging for privacy

```yaml
server:
  listen_address: "0.0.0.0:53"
  protocol: "udp"

upstream:
  servers:
    - "1.1.1.1:53" # Privacy-focused DNS (Cloudflare)
    - "9.9.9.9:53" # Privacy-focused DNS (Quad9)
  timeout: 3

blocklist:
  sources: ["blocklists/ads.txt"]

logging:
  level: "error" # Only errors
  log_queries: false # No query logging
  log_blocked: false # No blocked logging
  output_file: ""
```

---

### Testing Configuration

**Purpose:** Aggressive logging for debugging

```yaml
server:
  listen_address: "127.0.0.1:5353"
  protocol: "udp"

upstream:
  servers: ["8.8.8.8:53"]
  timeout: 5 # Longer timeout for debugging

blocklist:
  sources: ["blocklists/test-blocklist.txt"]

logging:
  level: "debug" # Maximum verbosity
  log_queries: true
  log_blocked: true
  output_file: "logs/debug.log"
```

---

## Updated API Reference

### Configuration (config.go)

**Types:**

```go
type Config struct {
    Server    ServerConfig
    Upstream  UpstreamConfig
    Blocklist BlocklistConfig
    Logging   LoggingConfig
}
```

**Functions:**

```go
// Create default configuration
func DefaultConfig() *Config

// Load from YAML file
func LoadFromFile(filepath string) (*Config, error)

// Validate configuration
func (c *Config) Validate() error
```

---

### Logger (logger.go)

**Types:**

```go
type Logger struct {
    // Internal fields
}

type LogLevel int
const (
    DEBUG LogLevel = iota
    INFO
    WARN
    ERROR
)
```

**Functions:**

```go
// Create new logger
func NewLogger(levelStr string, outputFile string) (*Logger, error)

// Logging methods
func (l *Logger) Debug(format string, args ...interface{})
func (l *Logger) Info(format string, args ...interface{})
func (l *Logger) Warn(format string, args ...interface{})
func (l *Logger) Error(format string, args ...interface{})

// Specialized query logging
func (l *Logger) Query(domain string, blocked bool, clientIP string)
```

---

### Statistics (stats.go)

**Types:**

```go
type Stats struct {
    // Thread-safe fields
}

type StatsSnapshot struct {
    Uptime         time.Duration
    TotalQueries   int64
    BlockedQueries int64
    AllowedQueries int64
    Errors         int64
    BlockedPercent float64
    QueriesPerSec  float64
}
```

**Functions:**

```go
// Create new stats tracker
func NewStats() *Stats

// Record events
func (s *Stats) RecordQuery(blocked bool)
func (s *Stats) RecordError()

// Get current statistics
func (s *Stats) GetStats() StatsSnapshot

// Format stats as string
func (ss StatsSnapshot) String() string
```

---

### Updated Server (server.go)

**Updated Constructor:**

```go
func NewDNSServer(cfg *config.Config, log *logger.Logger) *DNSServer
```

**New Method:**

```go
func (s *DNSServer) GetStats() stats.StatsSnapshot
```

---

## Testing Phase 3

### Unit Tests

**Test configuration loading:**

```go
// internal/config/config_test.go
func TestLoadConfig(t *testing.T) {
    cfg, err := LoadFromFile("testdata/valid.yaml")
    if err != nil {
        t.Fatal(err)
    }
    if cfg.Server.ListenAddress != "0.0.0.0:53" {
        t.Error("Wrong listen address")
    }
}

func TestInvalidConfig(t *testing.T) {
    cfg := &Config{
        Server: ServerConfig{
            ListenAddress: "", // Invalid
        },
    }
    if err := cfg.Validate(); err == nil {
        t.Error("Should fail validation")
    }
}
```

**Test logger:**

```go
// internal/logger/logger_test.go
func TestLogLevels(t *testing.T) {
    log, _ := NewLogger("warn", "")

    // Debug and Info should not appear
    log.Debug("This should not appear")
    log.Info("This should not appear")

    // Warn and Error should appear
    log.Warn("This should appear")
    log.Error("This should appear")
}
```

**Test statistics:**

```go
// internal/stats/stats_test.go
func TestStatsTracking(t *testing.T) {
    s := NewStats()

    s.RecordQuery(true)  // Blocked
    s.RecordQuery(false) // Allowed
    s.RecordQuery(true)  // Blocked

    stats := s.GetStats()

    if stats.TotalQueries != 3 {
        t.Errorf("Expected 3 total, got %d", stats.TotalQueries)
    }
    if stats.BlockedQueries != 2 {
        t.Errorf("Expected 2 blocked, got %d", stats.BlockedQueries)
    }
    if stats.AllowedQueries != 1 {
        t.Errorf("Expected 1 allowed, got %d", stats.AllowedQueries)
    }

    expectedPercent := (2.0 / 3.0) * 100.0
    if stats.BlockedPercent != expectedPercent {
        t.Errorf("Expected %.1f%% blocked, got %.1f%%",
                 expectedPercent, stats.BlockedPercent)
    }
}
```

### Integration Tests

**Test complete flow:**

```bash
# Start server with test config
sudo ./godnsblock --config test-config.yaml &
SERVER_PID=$!

# Wait for startup
sleep 2

# Test blocked domain
RESULT=$(dig @localhost ads.google.com +short)
if [ "$RESULT" != "0.0.0.0" ]; then
    echo "FAIL: Expected 0.0.0.0, got $RESULT"
fi

# Test allowed domain
RESULT=$(dig @localhost github.com +short)
if [ -z "$RESULT" ]; then
    echo "FAIL: Expected IP address, got nothing"
fi

# Stop server
kill $SERVER_PID

echo "Tests complete"
```

---

## Performance Considerations

### Memory Usage

**Phase 3 Additions:**

- Config structure: ~1KB
- Logger buffers: ~4KB per log
- Statistics: ~200 bytes

**Total overhead:** Minimal (~10-20KB additional)

### CPU Impact

**Logging:**

- File I/O: ~10-50μs per log entry
- String formatting: ~5-10μs
- Mitigation: Use appropriate log levels in production

**Statistics:**

- Mutex locking: ~50-100ns per query
- Counter increments: ~5-10ns
- Impact: Negligible (<1% overhead)

### Recommendations

**Development:**

- Use DEBUG level
- Log all queries
- Console output

**Production:**

- Use WARN or ERROR level
- Disable query logging
- Use file output with log rotation

---

## Troubleshooting Phase 3

### Config File Issues

**Problem:** "Failed to parse YAML"

```
Solution:
- Check YAML syntax (indentation with spaces, not tabs)
- Validate with: yamllint config.yaml
- Common issue: Missing quotes around strings with colons
```

**Problem:** "Invalid configuration: server.listen_address cannot be empty"

```
Solution:
- Ensure listen_address is set in config
- Check for typos in field names
- Verify YAML structure matches expected format
```

### Logger Issues

**Problem:** No logs appearing

```
Solution:
- Check log level (might be too high)
- Verify output_file path exists
- Check file permissions
- Ensure logging.log_queries = true for query logs
```

**Problem:** "Failed to open log file: permission denied"

```
Solution:
- Create log directory: mkdir -p logs
- Fix permissions: chmod 755 logs
- Run with appropriate privileges
```

### Statistics Issues

**Problem:** Statistics show 0 QPS but queries are working

```
Solution:
- Check if statistics are being recorded in handleQuery
- Verify RecordQuery() is called for both blocked and allowed
- Wait at least 60 seconds for first stats output
```

---

## Migration from Phase 2 to Phase 3

If you have Phase 2 code, here's how to upgrade:

### Step 1: Add Dependencies

```bash
go get gopkg.in/yaml.v3
```

### Step 2: Create New Packages

```bash
mkdir -p internal/config
mkdir -p internal/logger
mkdir -p internal/stats
mkdir -p configs
```

### Step 3: Update Imports in server.go

```go
import (
    // Add these:
    "github.com/yourusername/godnsblock/internal/config"
    "github.com/yourusername/godnsblock/internal/logger"
    "github.com/yourusername/godnsblock/internal/stats"
)
```

### Step 4: Update DNSServer Constructor

```go
// Old:
func NewDNSServer(address string, upstreamServer string) *DNSServer

// New:
func NewDNSServer(cfg *config.Config, log *logger.Logger) *DNSServer
```

### Step 5: Update main.go

- Add flag parsing
- Load configuration
- Initialize logger
- Pass to server constructor

### Step 6: Create config.yaml

- Use sample from documentation
- Customize for your needs

---

## Phase 3 Summary

### Achievements ✅

**Configuration Management:**

- Flexible YAML configuration
- Command-line overrides
- Validation and defaults
- Multiple environment support

**Observability:**

- Structured logging with levels
- Query tracking
- Real-time statistics
- Performance metrics

**Professional Features:**

- CLI interface
- Version information
- Graceful shutdown
- Statistics reporting

### Code Quality Improvements

**Before Phase 3:**

```go
// Hardcoded
listenAddr := "0.0.0.0:53"
upstreamDNS := "8.8.8.8:53"
log.Printf("Query for: %s", domain)
```

**After Phase 3:**

```go
// Configurable
cfg, _ := config.LoadFromFile("config.yaml")
logger.Info("Configuration loaded")
logger.Query(domain, blocked, clientIP)
stats.RecordQuery(blocked)
```

### Benefits

1. **Flexibility** - Easy configuration changes
2. **Observability** - Know what's happening
3. **Debuggability** - Detailed logs when needed
4. **Professionalism** - Production-ready features
5. **Maintainability** - Clean, organized code

---

## What's Next: Phase 4 Preview

Phase 4 will add advanced features:

1. **DNS Response Caching**

   - Cache resolved domains
   - TTL-based expiration
   - Memory management
   - Improved performance

2. **Multiple Upstream Resolvers**

   - Round-robin selection
   - Failover support
   - Health checking
   - Load balancing

3. **Remote Blocklist Support**

   - Download from URLs
   - Auto-update on schedule
   - Multiple sources
   - Merge deduplication

4. **Whitelist Support**
   - Override blocklist
   - Critical domain protection
   - User customization

---

**Phase 3 Complete! Your DNS ad blocker is now configurable, observable, and production-ready!** 🎉

---

**Document Version:** 1.0
**Last Updated:** February 13, 2024
**Project Version:** v0.3.0
