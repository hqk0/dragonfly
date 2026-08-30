package server

import (
	"net"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestAllowSupportedVersion(t *testing.T) {
	tests := []struct {
		version string
		allowed bool
	}{
		{version: "1.26.40", allowed: true},
		{version: "1.26.41", allowed: true},
		{version: "1.26.42", allowed: true},
		{version: "1.26.43", allowed: true},
		{version: "1.26.44", allowed: false},
		{version: "1.26.45", allowed: true},
	}
	for _, test := range tests {
		t.Run(test.version, func(t *testing.T) {
			next := &recordingAllower{reason: "allowed by configured allower", allowed: true}
			reason, allowed := allowSupportedVersion(next)(nil, login.IdentityData{}, login.ClientData{GameVersion: test.version})
			if allowed != test.allowed {
				t.Fatalf("allowed = %t, want %t (reason %q)", allowed, test.allowed, reason)
			}
			if test.version == "1.26.44" {
				if next.called {
					t.Fatal("configured allower was called for unsupported 1.26.44 client")
				}
				if reason == "" {
					t.Fatal("unsupported 1.26.44 client received an empty disconnect reason")
				}
				return
			}
			if !next.called {
				t.Fatal("configured allower was not called for supported client")
			}
			if reason != next.reason {
				t.Fatalf("reason = %q, want configured allower reason %q", reason, next.reason)
			}
		})
	}
}

func TestAllowSupportedVersionPreservesConfiguredRejection(t *testing.T) {
	next := &recordingAllower{reason: "banned", allowed: false}
	reason, allowed := allowSupportedVersion(next)(nil, login.IdentityData{}, login.ClientData{GameVersion: "1.26.45"})
	if allowed || reason != "banned" {
		t.Fatalf("Allow() = (%q, %t), want configured rejection", reason, allowed)
	}
}

type recordingAllower struct {
	called  bool
	reason  string
	allowed bool
}

func (a *recordingAllower) Allow(net.Addr, login.IdentityData, login.ClientData) (string, bool) {
	a.called = true
	return a.reason, a.allowed
}
