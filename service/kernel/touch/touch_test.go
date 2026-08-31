package touch

import (
	"reflect"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestServerRawsToServersIncludesFilterAndDisableMetadata(t *testing.T) {
	servers := serverRawsToServers([]configure.ServerRaw{
		{
			ServerObj: &serverObj.V2Ray{
				Ps:       "VMess TLS",
				Add:      "proxy.example",
				Port:     "443",
				Protocol: "vmess",
				Security: "aes-128-gcm",
				TLS:      "tls",
			},
		},
		{
			ServerObj: &serverObj.V2Ray{
				Ps:       "VLESS TLS none",
				Add:      "vless.example",
				Port:     "443",
				Protocol: "vless",
				TLS:      "tls",
			},
		},
	})

	if len(servers) != 2 {
		t.Fatalf("server count = %d, want 2", len(servers))
	}
	if got, want := servers[0].Protocol, "vmess"; got != want {
		t.Fatalf("protocol = %q, want %q", got, want)
	}
	if got, want := servers[0].Encryptions, []string{"tls", "aes-128-gcm"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("encryptions = %v, want %v", got, want)
	}
	if servers[1].DisableReason == "" {
		t.Fatal("TLS VLESS server with protocol encryption none has no disable reason")
	}
}
