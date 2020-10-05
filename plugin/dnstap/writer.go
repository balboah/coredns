package dnstap

import (
	"time"

	"github.com/coredns/coredns/plugin/dnstap/msg"
	tap "github.com/dnstap/golang-dnstap"
	"github.com/miekg/dns"
)

// ResponseWriter captures the client response and logs the query to dnstap.
// Single request use.
type ResponseWriter struct {
	QueryTime time.Time
	Query     *dns.Msg
	dns.ResponseWriter
	Tapper

	Err error
}

// WriteMsg writes back the response to the client and THEN works on logging the request
// and response to dnstap.
func (w *ResponseWriter) WriteMsg(resp *dns.Msg) error {
	writeErr := w.ResponseWriter.WriteMsg(resp)

	b := new(tap.Message)
	msg.SetResponseTime(b, time.Now())
	msg.SetQueryTime(b, w.QueryTime)
	if err := msg.SetQueryAddress(b, w.RemoteAddr()); err != nil {
		w.Err = err
		return err
	}

	if w.Pack() {
		buf, err := w.Query.Pack()
		if err != nil {
			w.Err = err
			return err
		}
		b.QueryMessage = buf
	}
	msg.SetType(b, tap.Message_CLIENT_QUERY)
	w.TapMessage(b)

	if writeErr != nil {
		return writeErr
	}

	if w.Pack() {
		buf, err := resp.Pack()
		if err != nil {
			w.Err = err
			return err
		}
		b.ResponseMessage = buf
	}
	msg.SetType(b, tap.Message_CLIENT_RESPONSE)
	w.TapMessage(b)

	return writeErr
}
