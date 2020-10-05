package forward

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/dnstap/msg"
	"github.com/coredns/coredns/request"

	tap "github.com/dnstap/golang-dnstap"
	"github.com/miekg/dns"
)

func toDnstap(ctx context.Context, host string, f *Forward, state request.Request, reply *dns.Msg, start time.Time) error {
	tapper := dnstap.TapperFromContext(ctx)
	if tapper == nil {
		return nil
	}
	// Query
	b := new(tap.Message)
	msg.SetQueryTime(b, start)
	ip, p, _ := net.SplitHostPort(host)     // this is preparsed and can't err here
	port, _ := strconv.ParseUint(p, 10, 32) // same here

	opts := f.opts
	t := state.Proto()
	switch {
	case opts.forceTCP: // TCP flag has precedence over UDP flag
		t = "tcp"
	case opts.preferUDP:
		t = "udp"
	}

	if t == "tcp" {
		ta := &net.TCPAddr{IP: net.ParseIP(ip), Port: int(port)}
		msg.SetQueryAddress(b, ta)
	} else {
		ta := &net.UDPAddr{IP: net.ParseIP(ip), Port: int(port)}
		msg.SetQueryAddress(b, ta)
	}

	if tapper.Pack() {
		buf, err := state.Req.Pack()
		if err != nil {
			return err
		}
		b.QueryMessage = buf
	}
	msg.SetType(b, tap.Message_FORWARDER_QUERY)
	tapper.TapMessage(b)

	// Response
	if reply != nil {
		if tapper.Pack() {
			buf, err := reply.Pack()
			if err != nil {
				return err
			}
			b.ResponseMessage = buf
		}
		msg.SetResponseTime(b, time.Now())
		msg.SetType(b, tap.Message_FORWARDER_RESPONSE)
		tapper.TapMessage(b)
	}

	return nil
}
