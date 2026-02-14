package server

import (
	"adblocker/internal/blocklist"
	"adblocker/internal/config"
	"adblocker/internal/logger"
	"adblocker/internal/resolver"
	"adblocker/internal/stats"
	"log"
	"net"

	"github.com/miekg/dns"
)

type DNSServer struct {
	config *config.Config
	server *dns.Server // poiner to dns.Server
	resolver *resolver.UpstreamResolver // pointer to resolver
	blocklist *blocklist.Blocklist
	logger *logger.Logger
	stats *stats.Stats
}

func NewDNSServer(cfg *config.Config, log *logger.Logger) *DNSServer {
	return &DNSServer{
		config: cfg,
		resolver: resolver.NewUpstreamResolver(cfg.Upstream.Servers[0]),
		blocklist: blocklist.NewBlocklist(),
		logger: log,
		stats: stats.NewStats(),
	}
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

	log.Printf("Starting DNS server on %s", s.config.Server.ListenAddress)

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

	

	log.Printf("Received query from %s", w.RemoteAddr())

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

	// Check Blocklist - This is the key part
	if s.blocklist.IsBlocked(domain) {
		log.Printf("BLOCKED: %s", domain)

		// Log blocked query
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

	// Record stats
	s.stats.RecordQuery(false)

	response, err := s.resolver.Resolve(r)
	if err != nil {
		s.logger.Error("Failed to resolve %s: %v", domain, err)
		s.stats.RecordError()
		//send error response
		dns.HandleFailed(w, r)
		return
	}

	w.WriteMsg(response)
}

func (s *DNSServer) GetStats() stats.StatsSnapshot {
	return s.stats.GetStats()
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