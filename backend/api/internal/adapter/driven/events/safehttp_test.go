package events

import (
	"net/netip"
	"testing"
)

func TestAllowedDestination(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "::1", "10.1.2.3", "172.16.0.1", "192.168.1.1", "169.254.169.254", "0.0.0.0",
		"100.64.0.1", "224.0.0.1", "fc00::1", "fe80::1", "::ffff:127.0.0.1", "::ffff:10.0.0.1", "240.0.0.1",
	}
	allowed := []string{"93.184.216.34", "8.8.8.8", "2606:4700:4700::1111"}
	for _, s := range blocked {
		if allowedDestination(netip.MustParseAddr(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	for _, s := range allowed {
		if !allowedDestination(netip.MustParseAddr(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
}

func TestGuardDial(t *testing.T) {
	if err := guardDial("tcp", "169.254.169.254:80", nil); err == nil {
		t.Error("metadata address must be refused")
	}
	if err := guardDial("tcp", "93.184.216.34:443", nil); err != nil {
		t.Errorf("public address refused: %v", err)
	}
}
