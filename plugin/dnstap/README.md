# dnstap

## Name

*dnstap* - enables logging to dnstap.

## Description

dnstap is a flexible, structured binary log format for DNS software; see https://dnstap.info. With this
plugin you make CoreDNS output dnstap logging.

Note that there is an internal buffer, so expect at least 13 requests before the server sends its
dnstap messages to the socket.

## Syntax

~~~ txt
dnstap SOCKET [full]
~~~

* **SOCKET** is the socket path supplied to the dnstap command line tool.
* `full` to include the wire-format DNS message.

## Examples

Log information about client requests and responses to */tmp/dnstap.sock*.

~~~ txt
dnstap /tmp/dnstap.sock
~~~

Log information including the wire-format DNS message about client requests and responses to */tmp/dnstap.sock*.

~~~ txt
dnstap unix:///tmp/dnstap.sock full
~~~

Log to a remote endpoint.

~~~ txt
dnstap tcp://127.0.0.1:6000 full
~~~

## Command Line Tool

Dnstap has a command line tool that can be used to inspect the logging. The tool can be found
at Github: <https://github.com/dnstap/golang-dnstap>. It's written in Go.

The following command listens on the given socket and decodes messages to stdout.

~~~ sh
$ dnstap -u /tmp/dnstap.sock
~~~

The following command listens on the given socket and saves message payloads to a binary dnstap-format log file.

~~~ sh
$ dnstap -u /tmp/dnstap.sock -w /tmp/test.dnstap
~~~

Listen for dnstap messages on port 6000.

~~~ sh
$ dnstap -l 127.0.0.1:6000
~~~

## Using Dnstap in your plugin

~~~ go
import (
    "github.com/coredns/coredns/plugin/dnstap"
    "github.com/coredns/coredns/plugin/dnstap/msg"

    tap "githum.com/dnstap/golang-dnstap"
)

func (h Dnstap) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
    // log client query to Dnstap
    if t := dnstap.TapperFromContext(ctx); t != nil {
        b := new(msg.Msg)
        msg.SetQueryTime(b, time.Now())
        msg.SetQueryAddress(b, w.RemoteAddr())
        if t.Pack() {
            buf, err := r.Pack()
            if err ! {
                return dns.RcodeServerFailure, err
            }
            b.QueryMessage = buf
        }
        msg.SetType(b, tap.Message_CLIENT_QUERY)
        t.TapMessage(b)
    }
    // ...
}
~~~

## See Also

The website [dnstap.info](https://dnstap.info) has info on the dnstap protocol.
