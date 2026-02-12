package resolver

import (
	"time"

	"github.com/miekg/dns"
)


type UpstreamResolver struct {
	server string
	client *dns.Client
}

func NewUpstreamResolver(server string) *UpstreamResolver {
	return &UpstreamResolver{
		server: server,
		client: &dns.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// resolve forwards a query to upstream DNS
func (u *UpstreamResolver) Resolve(query *dns.Msg) (*dns.Msg, error) {
	response, _, err := u.client.Exchange(query, u.server)

	if err != nil {
		return nil, err
	}

	return response, nil
}