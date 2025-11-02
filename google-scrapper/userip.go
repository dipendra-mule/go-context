package userip

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

func FromRequest(r *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", r.RemoteAddr)
	}
	return userIP, nil
}

// the key type is unexported to prevent collision with context keys defined in other pacakges
type key int

const userIPKey key = 0

func NewContext(ctx context.Context) (net.IP, bool) {
	return context.WithValue(ctx, userIPKey, userIP)
}

func FromContext(ctx context.Context) (net.IP, bool) {
	userIP, ok := ctx.Value(userIPKey).(net.IP)
	return userIP, ok
}
