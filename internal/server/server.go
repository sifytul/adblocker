package server

import (
	"adblocker/internal/blocklist"
	"adblocker/internal/cache"
	"adblocker/internal/config"
	"adblocker/internal/logger"
	"adblocker/internal/resolver"
	"adblocker/internal/stats"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

type DNSServer struct {
	config *config.Config
	server *dns.Server // poiner to dns.Server
	resolver *resolver.UpstreamResolver // pointer to resolver
	blocklist *blocklist.Blocklist
	cache *cache.Cache
	logger *logger.Logger
	stats *stats.Stats
}

func NewDNSServer(cfg *config.Config, log *logger.Logger) *DNSServer {
	server := &DNSServer{
		config: cfg,
		resolver: resolver.NewUpstreamResolver(cfg.Upstream.Servers[0]),
		blocklist: blocklist.NewBlocklist(),
		logger: log,
		stats: stats.NewStats(),
	}

	if cfg.Cache.Enabled {
		server.cache = cache.NewCache()
		log.Info("Cache enabled (cleanup: %ds)", cfg.Cache.CleanInterval)
	}

	return server
}

// method to load blocklist
func (s *DNSServer) LoadBlocklist(filepath string) error {
	return s.blocklist.LoadFromFile(filepath)
}

// Start begins listening for DNS queries on the specified address.
func (s *DNSServer) Start() error {
	// create DNS server that handles UDP
	s.server = &dns.Server{
		Addr: s.config.Server.ListenAddress,
		Net: "udp",
		Handler: dns.HandlerFunc(s.handleQuery),
	}

	s.logger.Info("Starting DNS server on %s", s.config.Server.ListenAddress)

	// Start cache cleanup goroutine if cache is enabled
	if s.config.Cache.Enabled && s.cache != nil {
		go s.cacheCleanupLoop()
	}


	// ListenAndServe blocks, so run in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %s", err)
		}
	}()

	return  nil
}

// Stop gracefully shuts down the Server
func (s *DNSServer) Stop() error {

	if s.server != nil {
		return s.server.Shutdown()
	}
	return nil
}

// handleQuery processes individual DNS queries
func (s *DNSServer) handleQuery(w dns.ResponseWriter, r *dns.Msg) {
	// r (request) contains the DNS query
	// w (writer) is used to send response


	if len(r.Question) == 0 {
		log.Printf("Empty query received")
		return
	}

	// Get the domain being queried
	question := r.Question[0]
	domain := question.Name
	qtype := question.Qtype // Query type (A, AAAA, CNAME etc.)
	clientIP := w.RemoteAddr().String()

	log.Printf("Query for domain: %s, type: %s", domain, dns.TypeToString[qtype])

	// 1. Check Blocklist
	if s.blocklist.IsBlocked(domain) {
		if s.config.Logging.LogBlocked {
			s.logger.Query(domain, true, clientIP)
		}

		// Record stats
		s.stats.RecordQuery(true)


		// create response with 0.0.0.0
		response := s.createBlockedResponse(r)
		w.WriteMsg(response)
		return
	}

	// Log allowed query
	if s.config.Logging.LogQueries {
		s.logger.Query(domain, false, clientIP)
	}

	// 2. Check cache if enabled
	if s.config.Cache.Enabled && s.cache != nil {
		if cachedResponse, found := s.cache.Get(domain, qtype); found {
			s.logger.Debug("Cache HIT: %s (type: %s)", domain, dns.TypeToString[qtype])

			// Update message ID to match the query
			cachedResponse.Id = r.Id

			w.WriteMsg(cachedResponse)

			if s.config.Logging.LogQueries {
				s.logger.Query(domain, false, clientIP)
			}
			s.stats.RecordQuery(false)
			return
		}

		s.logger.Debug("Cache MISS: %s (type: %s)", domain, dns.TypeToString[qtype])
	}

	// 3. Query upstream
	if s.config.Logging.LogQueries {
		s.logger.Query(domain, false, clientIP)
	}
	s.stats.RecordQuery(false)

	response, err := s.resolver.Resolve(r)
	if err != nil {
		s.logger.Error("Failed to resolve %s: %v", domain, err)
		s.stats.RecordError()
		dns.HandleFailed(w, r)
		return
	}

	// 4. Store in cache if enabled
	if s.config.Cache.Enabled && s.cache != nil {
		s.cache.Add(domain, qtype, response)
		
		// Log TTL for debugging
		if len(response.Answer) > 0 {
			ttl := response.Answer[0].Header().Ttl
			s.logger.Debug("Cached: %s (TTL: %ds)", domain, ttl)
		}
	}

	w.WriteMsg(response)
}

func (s *DNSServer) GetStats() stats.StatsSnapshot {
	return s.stats.GetStats()
}

func (s *DNSServer) GetCacheStats() cache.CacheStats {
	if s.cache != nil {
		return s.cache.Stats()
	}
	return cache.CacheStats{}
}

// createBlockedResponse returns a DNS response with 0.0.0.0
func (s *DNSServer) createBlockedResponse(request *dns.Msg) *dns.Msg {
	response := new(dns.Msg)
	response.SetReply(request)

	// Get the question
	question := request.Question[0]

	if question.Qtype == dns.TypeA {
		rr := &dns.A{
			Hdr: dns.RR_Header{
				Name: question.Name,
				Rrtype: dns.TypeA,
				Class: dns.ClassINET,
				Ttl: 300,
			},
			A: net.ParseIP("0.0.0.0"),
		}
		response.Answer = append(response.Answer, rr)
	}
	return response
}

// cacheCleanupLoop periodically removes expired cache entries
func (s *DNSServer) cacheCleanupLoop() {
	interval := time.Duration(s.config.Cache.CleanInterval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	s.logger.Debug("Cache cleanup started (interval: %s)", interval)

	for range ticker.C {
		removed := s.cache.CleanExpired()
		if removed > 0 {
			s.logger.Debug("Cache cleanup: removed %d expired entries", removed)
		}
	}
}