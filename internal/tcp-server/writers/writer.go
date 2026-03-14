package writers

import (
	"encoding/binary"
	"fmt"
	"net"
	"transaction-monitoring-system/protobuf"

	"google.golang.org/protobuf/proto"
)

type Writer struct{}

func (h *Writer) WriteResponse(conn net.Conn, payload []byte) error {
	const op = "internal.tcp-server.writers.WriteResponse"
	resp := &protobuf.Response{
		Ok:     true,
		Result: payload,
	}

	if err := h.WriteMessage(conn, resp); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (h *Writer) WriteError(conn net.Conn, msg string) error {
	const op = "internal.tcp-server.writers.WriteError"

	resp := &protobuf.Response{
		Ok:    false,
		Error: msg,
	}

	if err := h.WriteMessage(conn, resp); err != nil {
		return fmt.Errorf("%s ; %s", op, err)
	}

	return nil
}

func (h *Writer) WriteMessage(conn net.Conn, message proto.Message) error {
	const op = "internal.tcp-server.writers.WriteMessage"

	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}

	lenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBuf, uint32(len(data)))

	if _, err = conn.Write(lenBuf); err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}
	if _, err = conn.Write(data); err != nil {
		return fmt.Errorf("%s : %s", op, err)
	}

	return nil
}
