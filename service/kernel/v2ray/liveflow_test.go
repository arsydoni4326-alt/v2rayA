package v2ray

import "testing"

func TestParseLiveFlowEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		valid    bool
		ip       string
		port     int
		protocol string
	}{
		{name: "tcp ipv4", raw: "tcp:192.0.2.10:443", valid: true, ip: "192.0.2.10", port: 443, protocol: "tcp"},
		{name: "udp ipv6", raw: "udp:[2001:db8::1]:53", valid: true, ip: "2001:db8::1", port: 53, protocol: "udp"},
		{name: "http url", raw: "https://example.com:8443/path", valid: true, ip: "example.com", port: 8443},
		{name: "empty", raw: "", valid: false},
		{name: "bad port", raw: "tcp:example.com:70000", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			endpoint, protocol, ok := parseLiveFlowEndpoint(test.raw)
			if ok != test.valid {
				t.Fatalf("validity = %v, want %v", ok, test.valid)
			}
			if !test.valid {
				return
			}
			if endpoint.IP != test.ip || endpoint.Port != test.port || protocol != test.protocol {
				t.Fatalf("endpoint = %#v, protocol = %q", endpoint, protocol)
			}
		})
	}
}

func TestSelectedDetourTag(t *testing.T) {
	tests := map[string]string{
		"":                      "",
		"proxy":                 "proxy",
		"socks -> 『Tokyo-01』":   "Tokyo-01",
		"socks ==> 『Tokyo-01』":  "Tokyo-01",
		"transparent >> direct": "direct",
	}
	for input, want := range tests {
		if got := selectedDetourTag(input); got != want {
			t.Errorf("selectedDetourTag(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestConsumeCoreLiveFlowEvent(t *testing.T) {
	if !ConsumeCoreLiveFlowEvent(`V2RAYA_LIVE_FLOW {"source":"tcp:192.0.2.1:1111","destination":"tcp:example.com:443","detour":"direct","timestamp":1}`) {
		t.Fatal("valid live flow event was not consumed")
	}
	if !ConsumeCoreLiveFlowEvent("V2RAYA_LIVE_FLOW not-json") {
		t.Fatal("malformed private event was not consumed")
	}
	if ConsumeCoreLiveFlowEvent("ordinary core log") {
		t.Fatal("ordinary core log was consumed")
	}
}
