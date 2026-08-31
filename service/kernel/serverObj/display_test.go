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
			disabled: false,
		},
		{
			name:     "VLESS TLS WS on domain",
			server:   &V2Ray{Protocol: "vless", TLS: "tls", Net: "ws", Add: "indo8.vpnjantit.com:10002"},
			disabled: false,
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
			server:   &V2Ray{Protocol: "vless", TLS: "", Add: "104.26.13.40:443"},
			disabled: true,
		},
		{
			name:     "VLESS empty TLS on domain",
			server:   &V2Ray{Protocol: "vless", TLS: "", Add: "proxy.example.com:443"},
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
			disabled: true,
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
		// security=false in VLESS URLs sets v.TLS to "false" — must be blocked
		{
			name:     "VLESS security=false on public IP",
			server:   &V2Ray{Protocol: "vless", TLS: "false", Add: "104.26.13.40:8880"},
			disabled: true,
		},
		{
			name:     "VLESS security=false on private IP",
			server:   &V2Ray{Protocol: "vless", TLS: "false", Add: "192.168.1.1:8880"},
			disabled: true,
		},
		{
			name:     "VLESS security=false on domain",
			server:   &V2Ray{Protocol: "vless", TLS: "false", Add: "proxy.example.com:8880"},
			disabled: true,
		},
		// security=none in VLESS URLs sets v.TLS to "none" — also blocked
		{
			name:     "VLESS security=none on domain",
			server:   &V2Ray{Protocol: "vless", TLS: "none", Add: "itproxy4.lockdwn.com:443"},
			disabled: true,
		},
		{
			name:     "VMess security=false encryption none",
			server:   &V2Ray{Protocol: "vmess", TLS: "false", Security: "none"},
			disabled: true,
		},
		{
			name:     "VMess security=false with real encryption",
			server:   &V2Ray{Protocol: "vmess", TLS: "false", Security: "aes-128-gcm"},
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
