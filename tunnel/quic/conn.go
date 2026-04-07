package quic

import (
"net"

"github.com/quic-go/quic-go"

"github.com/p4gefau1t/trojan-go/tunnel"
)

type Conn struct {
quic.Stream
conn quic.Connection
}

func (c *Conn) Read(p []byte) (n int, err error) {
return c.Stream.Read(p)
}

func (c *Conn) Write(p []byte) (n int, err error) {
return c.Stream.Write(p)
}

func (c *Conn) Close() error {
c.Stream.CancelRead(0)
return c.Stream.Close()
}

func (c *Conn) LocalAddr() net.Addr {
return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
return c.conn.RemoteAddr()
}

func (c *Conn) Metadata() *tunnel.Metadata {
return nil
}
