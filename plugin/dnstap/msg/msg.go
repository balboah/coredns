package msg

import (
	"fmt"
	"net"
	"time"

	tap "github.com/dnstap/golang-dnstap"
)

// SetQueryAddress adds the query address to the message. This also sets the SocketFamily and SocketProtocol.
func SetQueryAddress(t *tap.Message, addr net.Addr) error {
	inet := tap.SocketFamily_INET
	t.SocketFamily = &inet
	switch a := addr.(type) {
	case *net.TCPAddr:
		t.QueryAddress = a.IP
		p := uint32(a.Port)
		t.QueryPort = &p
		p1 := tap.SocketProtocol_TCP
		t.SocketProtocol = &p1
		if a.IP.To4() == nil {
			inet6 := tap.SocketFamily_INET6
			t.SocketFamily = &inet6
		}
		return nil
	case *net.UDPAddr:
		t.QueryAddress = a.IP
		p := uint32(a.Port)
		t.QueryPort = &p
		p1 := tap.SocketProtocol_UDP
		t.SocketProtocol = &p1
		if a.IP.To4() == nil {
			inet6 := tap.SocketFamily_INET6
			t.SocketFamily = &inet6
		}
		return nil
	default:
		return fmt.Errorf("unknown address type: %T", a)
	}
}

// SetResponseAddress the response address to the message.
func SetResponseAddress(t *tap.Message, addr net.Addr) error {
	switch a := addr.(type) {
	case *net.TCPAddr:
		t.ResponseAddress = a.IP
		p := uint32(a.Port)
		t.ResponsePort = &p
		return nil
	case *net.UDPAddr:
		t.ResponseAddress = a.IP
		p := uint32(a.Port)
		t.ResponsePort = &p
		return nil
	default:
		return fmt.Errorf("unknown address type: %T", a)
	}
}

// SetQueryTime sets the time of the query in t.
func SetQueryTime(t *tap.Message, ti time.Time) {
	qts := uint64(ti.Unix())
	qtn := uint32(ti.Nanosecond())
	t.QueryTimeSec = &qts
	t.QueryTimeNsec = &qtn
}

// SetResponseTime sets the time of the response in t.
func SetResponseTime(t *tap.Message, ti time.Time) {
	rts := uint64(ti.Unix())
	rtn := uint32(ti.Nanosecond())
	t.ResponseTimeSec = &rts
	t.ResponseTimeNsec = &rtn
}

// SetType sets the type in t.
func SetType(t *tap.Message, typ tap.Message_Type) { t.Type = &typ }
