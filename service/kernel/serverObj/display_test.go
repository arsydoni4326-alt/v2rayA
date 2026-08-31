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
			disabled: true,
		},
		{
			name:     "VLESS empty TLS",
			server:   &V2Ray{Protocol: "vless", TLS: ""},
			disabled: true,
		},
		{
			name:     "VMess empty TLS empty Security defaults to auto",
			server:   &V2Ray{Protocol: "vmess", TLS: "", Security: ""},
			disabled: false,
		},
		{
			name:     "VLESS empty TLS private address",
			server:   &V2Ray{Protocol: "vless", TLS: "", Add: "192.168.1.1:443"},
			disabled: false,
		},
		{
			name:     "VLESS Reality",
			server:   &V2Ray{Protocol: "vless", TLS: "reality", PublicKey: "abc123"},
			disabled: false,
		},
		{
			name:     "VMess without TLS with real encryption",
			server:   &V2Ray{Protocol: "vmess", TLS: "", Security: "aes-128-gcm"},
			disabled: false,
		},
		{
			name:     "VMess TLS none with cipher",
			server:   &V2Ray{Protocol: "vmess", TLS: "none", Security: "aes-128-gcm"},
			disabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reason := tt.server.DisableReason("test server", "proxy.example:443")
			if got := reason != ""; got != tt.disabled {
				t.Fatalf("DisableReason() = %q, disabled = %v, want disabled = %v", reason, got, tt.disabled)
			}
			if tt.disabled && !strings.Contains(reason, "is excluded") && !strings.Contains(reason, "is not supported") && !strings.Contains(reason, "is prohibited") {
				t.Fatalf("DisableReason() = %q, want an exclusion reason", reason)
			}
		})
	}
}
