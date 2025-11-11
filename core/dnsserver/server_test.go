package dnsserver

import (
	"context"
	"errors"
	"net"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

type testPlugin struct{}

func (tp testPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	return 0, nil
}

func (tp testPlugin) Name() string { return "local" }

// blockingPlugin uses sync.Mutex to simulate extended processing.
type blockingPlugin struct {
	sync.Mutex
}

func (b *blockingPlugin) Name() string { return "blocking" }

func (b *blockingPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Respond immediately to avoid waiting in dns.Exchange
	m := new(dns.Msg)
	m.SetRcodeFormatError(r)
	w.WriteMsg(m)

	b.Lock()
	defer b.Unlock()
	return dns.RcodeSuccess, nil
}

func testConfig(transport string, p plugin.Handler) *Config {
	c := &Config{
		Zone:        "example.com.",
		Transport:   transport,
		ListenHosts: []string{"127.0.0.1"},
		Port:        "53",
		Debug:       false,
		Stacktrace:  false,
	}

	c.AddPlugin(func(next plugin.Handler) plugin.Handler { return p })
	return c
}

func TestNewServer(t *testing.T) {
	_, err := NewServer("127.0.0.1:53", []*Config{testConfig("dns", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}

	_, err = NewServergRPC("127.0.0.1:53", []*Config{testConfig("grpc", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServergRPC, got %s", err)
	}

	_, err = NewServerTLS("127.0.0.1:53", []*Config{testConfig("tls", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServerTLS, got %s", err)
	}

	_, err = NewServerQUIC("127.0.0.1:53", []*Config{testConfig("quic", testPlugin{})})
	if err != nil {
		t.Errorf("Expected no error for NewServerQUIC, got %s", err)
	}
}

func TestDebug(t *testing.T) {
	configNoDebug, configDebug := testConfig("dns", testPlugin{}), testConfig("dns", testPlugin{})
	configDebug.Debug = true

	s1, err := NewServer("127.0.0.1:53", []*Config{configDebug, configNoDebug})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}
	if !s1.debug {
		t.Errorf("Expected debug mode enabled for server s1")
	}
	if !log.D.Value() {
		t.Errorf("Expected debug logging enabled")
	}

	s2, err := NewServer("127.0.0.1:53", []*Config{configNoDebug})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}
	if s2.debug {
		t.Errorf("Expected debug mode disabled for server s2")
	}
	if log.D.Value() {
		t.Errorf("Expected debug logging disabled")
	}
}

func TestStacktrace(t *testing.T) {
	configNoStacktrace, configStacktrace := testConfig("dns", testPlugin{}), testConfig("dns", testPlugin{})
	configStacktrace.Stacktrace = true

	s1, err := NewServer("127.0.0.1:53", []*Config{configStacktrace, configStacktrace})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}
	if !s1.stacktrace {
		t.Errorf("Expected stacktrace mode enabled for server s1")
	}

	s2, err := NewServer("127.0.0.1:53", []*Config{configNoStacktrace})
	if err != nil {
		t.Errorf("Expected no error for NewServer, got %s", err)
	}
	if s2.stacktrace {
		t.Errorf("Expected stacktrace disabled for server s2")
	}
}

func TestGracefulStopTimeout_Internal(t *testing.T) {
	p := new(blockingPlugin)
	cfg := testConfig("dns", p)

	s, err := NewServer("127.0.0.1:0", []*Config{cfg})
	if err != nil {
		t.Fatalf("NewServer failed: %v", err)
	}

	// Shorten the graceful timeout
	s.graceTimeout = 500 * time.Millisecond

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("ListenPacket failed: %v", err)
	}
	defer pc.Close()

	go s.ServePacket(pc)
	udp := pc.LocalAddr().String()

	// Block the handler
	p.Lock()
	defer p.Unlock()

	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)
	_, err = dns.Exchange(m, udp)
	if err != nil {
		t.Fatalf("dns.Exchange failed: %v", err)
	}

	err = s.Stop()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func BenchmarkCoreServeDNS(b *testing.B) {
	s, err := NewServer("127.0.0.1:53", []*Config{testConfig("dns", testPlugin{})})
	if err != nil {
		b.Errorf("Expected no error for NewServer, got %s", err)
	}

	ctx := context.TODO()
	w := &test.ResponseWriter{}
	m := new(dns.Msg)
	m.SetQuestion("aaa.example.com.", dns.TypeTXT)

	b.ReportAllocs()

	for b.Loop() {
		s.ServeDNS(ctx, w, m)
	}
}

func TestComposeControlFuncs(t *testing.T) {
	called := []int{}
	fst := func(network, address string, c syscall.RawConn) error {
		called = append(called, 1)
		return nil
	}
	snd := func(network, address string, c syscall.RawConn) error {
		called = append(called, 2)
		return nil
	}

	combined := composeControlFuncs([]ControlFunc{nil, fst, snd})
	if combined == nil {
		t.Fatalf("expected combined control func")
	}
	if err := combined("tcp", ":0", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(called) != 2 || called[0] != 1 || called[1] != 2 {
		t.Fatalf("control funcs called out of order: %v", called)
	}

	errSentinel := errors.New("boom")
	errCombined := composeControlFuncs([]ControlFunc{func(string, string, syscall.RawConn) error { return errSentinel }, snd})
	if err := errCombined("tcp", ":0", nil); !errors.Is(err, errSentinel) {
		t.Fatalf("expected %v, got %v", errSentinel, err)
	}

	if composeControlFuncs(nil) != nil {
		t.Fatalf("expected nil when no control funcs provided")
	}
}

func TestComposePacketConnSetups(t *testing.T) {
	called := []int{}
	fst := func(conn *net.UDPConn) error {
		called = append(called, 1)
		return nil
	}
	snd := func(conn *net.UDPConn) error {
		called = append(called, 2)
		return nil
	}

	combined := composePacketConnSetups([]PacketConnSetupFunc{nil, fst, snd})
	if combined == nil {
		t.Fatalf("expected combined packet setup func")
	}
	if err := combined(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(called) != 2 || called[0] != 1 || called[1] != 2 {
		t.Fatalf("packet setup funcs called out of order: %v", called)
	}

	errSentinel := errors.New("boom")
	errCombined := composePacketConnSetups([]PacketConnSetupFunc{func(*net.UDPConn) error { return errSentinel }, snd})
	if err := errCombined(nil); !errors.Is(err, errSentinel) {
		t.Fatalf("expected %v, got %v", errSentinel, err)
	}

	if composePacketConnSetups(nil) != nil {
		t.Fatalf("expected nil when no packet setup funcs provided")
	}
}
