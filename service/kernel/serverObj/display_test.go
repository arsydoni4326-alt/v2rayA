package serverObj

import (
	"reflect"
	"strings"
	"testing"
)

func TestEncryptionTypes(t *testing.T) {
	tests := []struct {
		name   string
		server ServerObj
		want   []string
	}{
		{
			name:   "VMess cipher and TLS",
			server: &V2Ray{Protocol: "vmess", Security: "aes-128-gcm", TLS: "tls"},
			want:   []string{"tls", "aes-128-gcm"},
		},
		{
			name:   "VLESS Reality",
			server: &V2Ray{Protocol: "vless", TLS: "reality"},
			want:   []string{"reality"},
		},
		{
			name:   "Shadowsocks cipher",
			server: &Shadowsocks{Cipher: "chacha20-ietf-poly1305"},
			want:   []string{"chacha20-ietf-poly1305"},
		},
		{
			name:   "HTTPS proxy",
			server: &HTTP{Protocol: "https"},
			want:   []string{"tls"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncryptionTypes(tt.server); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("EncryptionTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV2RayDisableReasonForTLSEncryptionNoneOrFalse(t *testing.T) {
	tests := []struct {
		name     string
		server   *V2Ray
		disabled bool
	}{
		{
			name:     "VMess TLS encryption none",
			server:   &V2Ray{Protocol: "vmess", TLS: "tls", Security: "none"},
			disabled: true,
		},
		{
			name:     "VMess TLS encryption false",
			server:   &V2Ray{Protocol: "vmess", TLS: "TLS", Security: "FALSE"},
			disabled: true,
		},
		{
			name:     "VLESS TLS protocol encryption none",
			server:   &V2Ray{Protocol: "vless", TLS: "tls"},
			disabled: true,
		},
		{
			name:     "VMess TLS encrypted",
			server:   &V2Ray{Protocol: "vmess", TLS: "tls", Security: "aes-128-gcm"},
			disabled: false,
		},
		{
			name:     "VMess without TLS encryption none",
			server:   &V2Ray{Protocol: "vmess", TLS: "none", Security: "none"},
			disabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := tt.server.DisableReason("test server", "proxy.example:443")
			if got := reason != ""; got != tt.disabled {
				t.Fatalf("DisableReason() = %q, disabled = %v, want disabled = %v", reason, got, tt.disabled)
			}
			if tt.disabled && !strings.Contains(reason, "TLS with encryption none or false") {
				t.Fatalf("DisableReason() = %q, want TLS encryption exclusion reason", reason)
			}
		})
	}
}
