package base_handler

import (
	"encoding/binary"
	"io"
	"log/slog"
	"net"
	"sync"
	"transaction-monitoring-system/protobuf"

	"google.golang.org/protobuf/proto"
)

type CustomHandler interface {
	Handle(conn net.Conn, req *protobuf.Request)
	Type() string
}

type Handler struct {
	log            *slog.Logger
	customHandlers map[string]CustomHandler
}

func NewHandler(log *slog.Logger, handlers ...CustomHandler) *Handler {
	register := make(map[string]CustomHandler)
	for _, h := range handlers {
		register[h.Type()] = h
	}

	return &Handler{
		log:            log,
		customHandlers: register,
	}
}

func (h *Handler) Handle(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		wg.Done()
	}()

	const op = "internal.tcp-server.base-handler.Handler.Handle"

	handlerlog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	handlerlog.Info("new client connected")

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			handlerlog.Error("something wrong with length prefix", slog.String("error", err.Error()))
			return
		}

		length := binary.BigEndian.Uint32(lenBuf)
		if length == 0 {
			handlerlog.Info("request body is empty")
			continue
		}
		if length > 4<<20 { // 4 MB
			handlerlog.Info("request body is too long")
			continue
		}

		message := make([]byte, length)
		_, err = io.ReadFull(conn, message)
		if err != nil {
			handlerlog.Error("something wrong with payload", slog.String("error", err.Error()))
			return
		}

		var req protobuf.Request
		if err = proto.Unmarshal(message, &req); err != nil {
			handlerlog.Error("bad unmarshal message", slog.String("error", err.Error()))
			continue
		}

		handler, exists := h.customHandlers[req.Type]
		if !exists {
			handlerlog.Error("custom handler not found", slog.String("type", req.Type))
			return
		}
		handler.Handle(conn, &req)
	}
}
