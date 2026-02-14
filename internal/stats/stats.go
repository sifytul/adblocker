package stats

import (
	"fmt"
	"sync"
	"time"
)

// Stats tracks DNS query statistics
type Stats struct {
    mu sync.RWMutex
    
    StartTime      time.Time
    TotalQueries   int64
    BlockedQueries int64
    AllowedQueries int64
    Errors         int64
}

// NewStats creates a new stats tracker
func NewStats() *Stats {
    return &Stats{
        StartTime: time.Now(),
    }
}

// RecordQuery records a DNS query
func (s *Stats) RecordQuery(blocked bool) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.TotalQueries++
    if blocked {
        s.BlockedQueries++
    } else {
        s.AllowedQueries++
    }
}

// RecordError records an error
func (s *Stats) RecordError() {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.Errors++
}

// GetStats returns a copy of current stats
func (s *Stats) GetStats() StatsSnapshot {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    uptime := time.Since(s.StartTime)
    
    return StatsSnapshot{
        Uptime:         uptime,
        TotalQueries:   s.TotalQueries,
        BlockedQueries: s.BlockedQueries,
        AllowedQueries: s.AllowedQueries,
        Errors:         s.Errors,
        BlockedPercent: s.calculateBlockedPercent(),
        QueriesPerSec:  s.calculateQPS(uptime),
    }
}

// StatsSnapshot is a read-only view of stats
type StatsSnapshot struct {
    Uptime         time.Duration
    TotalQueries   int64
    BlockedQueries int64
    AllowedQueries int64
    Errors         int64
    BlockedPercent float64
    QueriesPerSec  float64
}

// calculateBlockedPercent calculates percentage of blocked queries
func (s *Stats) calculateBlockedPercent() float64 {
    if s.TotalQueries == 0 {
        return 0.0
    }
    return (float64(s.BlockedQueries) / float64(s.TotalQueries)) * 100.0
}

// calculateQPS calculates queries per second
func (s *Stats) calculateQPS(uptime time.Duration) float64 {
    seconds := uptime.Seconds()
    if seconds == 0 {
        return 0.0
    }
    return float64(s.TotalQueries) / seconds
}

// String returns formatted stats
func (ss StatsSnapshot) String() string {
    return fmt.Sprintf(
        "Uptime: %s | Total: %d | Blocked: %d (%.1f%%) | Allowed: %d | Errors: %d | QPS: %.2f",
        ss.Uptime.Round(time.Second),
        ss.TotalQueries,
        ss.BlockedQueries,
        ss.BlockedPercent,
        ss.AllowedQueries,
        ss.Errors,
        ss.QueriesPerSec,
    )
}