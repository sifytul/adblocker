package server

import (
	"adblocker/internal/resolver"
	"log"

	"github.com/miekg/dns"
)

type DNSServer struct {
	address string     // Listen address (e.g., "0.0.0.0:53")
	server *dns.Server // poiner to dns.Server
	resolver *resolver.UpstreamResolver // pointer to resolver
}

func NewDNSServer(address string, upstreamServer string) *DNSServer {
	return &DNSServer{
		address: address,
		resolver: resolver.NewUpstreamResolver(upstreamServer),
	}
}


// Start begins listening for DNS queries on the specified address.
func (s *DNSServer) Start() error {
	// create DNS server that handles UDP
	s.server = &dns.Server{
		Addr: s.address,
		Net: "udp",
		Handler: dns.HandlerFunc(s.handleQuery),
	}

	log.Printf("Starting DNS server on %s", s.address)

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

	log.Printf("Query for domain: %s, type: %s", domain, dns.TypeToString[qtype])

	response, err := s.resolver.Resolve(r)
	if err != nil {
		log.Printf("Error resolving query %s: %v", domain, err)
		//send error response
		dns.HandleFailed(w, r)
		return
	}

	w.WriteMsg(response)
	log.Printf("Resolved %s successfully", domain)
}