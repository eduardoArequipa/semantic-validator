package ratelimit

import "testing"

func TestClientIP(t *testing.T) {
	tests := []struct {
		name              string
		remoteAddr        string
		forwardedFor      string
		trustProxyHeaders bool
		want              string
	}{
		{
			name:         "uses peer address by default",
			remoteAddr:   "10.0.0.2:4567",
			forwardedFor: "203.0.113.9",
			want:         "10.0.0.2",
		},
		{
			name:              "uses rightmost forwarded address from trusted proxy",
			remoteAddr:        "10.0.0.2:4567",
			forwardedFor:      "198.51.100.4, 203.0.113.9",
			trustProxyHeaders: true,
			want:              "203.0.113.9",
		},
		{
			name:              "ignores invalid forwarded address",
			remoteAddr:        "10.0.0.2:4567",
			forwardedFor:      "attacker-value",
			trustProxyHeaders: true,
			want:              "10.0.0.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientIP(tt.remoteAddr, tt.forwardedFor, tt.trustProxyHeaders); got != tt.want {
				t.Fatalf("clientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}
