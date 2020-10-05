package dnstap

import (
	"context"
	"time"

	"github.com/coredns/coredns/plugin"

	tap "github.com/dnstap/golang-dnstap"
	"github.com/miekg/dns"
)

// Dnstap is the dnstap handler.
type Dnstap struct {
	Next plugin.Handler
	IO   IORoutine

	// Set to true to include the relevant raw DNS message into the dnstap messages.
	JoinRawMessage bool
}

type (
	// IORoutine is the dnstap I/O thread as defined by: <http://dnstap.info/Architecture>.
	IORoutine interface {
		Dnstap(tap.Dnstap)
	}
	// Tapper is implemented by the Context passed by the dnstap handler.
	Tapper interface {
		TapMessage(message *tap.Message)
		Pack() bool
	}
)

// TapMessage implements Tapper.
func (h Dnstap) TapMessage(m *tap.Message) {
	t := tap.Dnstap_MESSAGE
	h.IO.Dnstap(tap.Dnstap{
		Type:    &t,
		Message: m,
	})
}

// Pack returns true if the raw DNS message should be included into the dnstap messages.
func (h Dnstap) Pack() bool {
	return h.JoinRawMessage
}

// ServeDNS logs the client query and response to dnstap and passes the dnstap Context.
func (h Dnstap) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	rw := &ResponseWriter{
		ResponseWriter: w,
		Tapper:         &h,
		Query:          r,
		QueryTime:      time.Now(),
	}

	code, err := plugin.NextOrFailure(h.Name(), h.Next, contextWithTapper(ctx, h), rw, r)
	if err != nil {
		return code, err
	}

	if rw.Err != nil {
		return code, plugin.Error("dnstap", rw.Err)
	}

	return code, nil
}

// Name returns dnstap.
func (h Dnstap) Name() string { return "dnstap" }
