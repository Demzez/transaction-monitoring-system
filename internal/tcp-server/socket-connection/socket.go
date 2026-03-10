package socket_connection

import (
	"io"
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/repository/postgres"
)

type Handler struct {
	db  *postgres.Repository
	log *slog.Logger
}

func NewHandler(log *slog.Logger, db *postgres.Repository) *Handler {
	return &Handler{
		db:  db,
		log: log,
	}
}

func (h *Handler) Handle(conn net.Conn) {
	defer conn.Close()

	const op = "internal.tcp-server.socket-connection.Handler.Handle"

	handlerlog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenBuf)
	if err != nil {
		handlerlog.Error("something wrong with length prefix", slog.String("error", err.Error()))
		return
	}

	// TODO: finish write TCP handler and start special handlers and responses writer
	
}
