package writers

import "net"

type WrInterface interface {
	WriteResponse(conn net.Conn, payload []byte) error
	WriteError(conn net.Conn, msg string) error
}
