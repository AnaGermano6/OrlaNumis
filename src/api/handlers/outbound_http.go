package handlers

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"

	"github.com/briandenicola/ancient-coins-api/services"
)

var errOutboundTargetBlocked = services.ErrOutboundTargetBlocked

type outboundResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

var defaultOutboundHTTPClient = newRestrictedHTTPClient(net.DefaultResolver)

var outboundHTTPClientFactory = func() *http.Client {
	return defaultOutboundHTTPClient
}

func validateOutboundURL(rawURL string) (*url.URL, error) {
	return services.ValidateOutboundURL(rawURL)
}

func newRestrictedHTTPClient(resolver outboundResolver) *http.Client {
	return services.NewRestrictedHTTPClient(resolver, 30*time.Second, 10)
}

func isOutboundTargetBlockedError(err error) bool {
	return services.IsOutboundTargetBlockedError(err)
}
