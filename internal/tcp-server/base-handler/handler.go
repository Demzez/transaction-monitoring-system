package base_handler

import (
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
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
	idleTimeout    time.Duration
}

func NewHandler(log *slog.Logger, idleTimeout time.Duration, handlers ...CustomHandler) *Handler {
	register := make(map[string]CustomHandler)
	for _, h := range handlers {
		register[h.Type()] = h
	}

	return &Handler{
		log:            log,
		customHandlers: register,
		idleTimeout:    idleTimeout,
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
	if h.idleTimeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(h.idleTimeout)); err != nil {
			handlerlog.Error("failed to set deadline", slog.String("error", err.Error()))
		}
	}

	for {
		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				handlerlog.Warn("timeout exceeded", slog.String("error", err.Error()))
			} else {
				handlerlog.Error("something wrong with length prefix", slog.String("error", err.Error()))
			}
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

// TODO : token validate
/*
if req.Token == "" {
        handlerlog.Error("missing token")
        _ = h.wr.WriteError(conn, "missing token") // use writer to send error
        return
    }
    // You need to inject the JWT secret into the handler. For simplicity, you can add it to Handler struct.
    // Let's add a field: jwtSecret string
    // Then validate:
    _, err = jwt.ValidateToken(req.Token, h.jwtSecret)
    if err != nil {
        handlerlog.Error("invalid token", slog.String("error", err.Error()))
        _ = h.wr.WriteError(conn, "invalid token")
        return
    }
*/
