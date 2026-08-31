package serverObj

import (
	"reflect"
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
