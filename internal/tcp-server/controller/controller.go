package controller

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

type Handler interface {
	Handle(conn net.Conn, req *protobuf.Request)
	Type() string
}

type Controller struct { // TODO: передать сюда конфиг для проверки токена && refactoring!!!
	log            *slog.Logger
	customHandlers map[string]Handler
	idleTimeout    time.Duration
}

func NewController(log *slog.Logger, idleTimeout time.Duration, handlers ...Handler) *Controller {
	register := make(map[string]Handler)
	for _, h := range handlers {
		register[h.Type()] = h
	}

	return &Controller{
		log:            log,
		customHandlers: register,
		idleTimeout:    idleTimeout,
	}
}

func (h *Controller) Process(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		wg.Done()
	}()

	const op = "internal.tcp-server.controller.handler.Process"

	controllerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	controllerLog.Info("new client connected")

	for {
		if h.idleTimeout > 0 {
			if err := conn.SetDeadline(time.Now().Add(h.idleTimeout)); err != nil {
				controllerLog.Error("failed to set deadline", slog.String("error", err.Error()))
			}
		}

		lenBuf := make([]byte, 4)
		_, err := io.ReadFull(conn, lenBuf)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				controllerLog.Warn("timeout exceeded", slog.String("error", err.Error()))
			} else {
				controllerLog.Error("something wrong with length prefix", slog.String("error", err.Error()))
			}
			return
		}

		length := binary.BigEndian.Uint32(lenBuf)
		if length == 0 {
			controllerLog.Info("request body is empty")
			continue
		}
		if length > 4<<20 { // 4 MB
			controllerLog.Info("request body is too long")
			continue
		}

		message := make([]byte, length)
		_, err = io.ReadFull(conn, message)
		if err != nil {
			controllerLog.Error("something wrong with payload", slog.String("error", err.Error()))
			return
		}

		var req protobuf.Request
		if err = proto.Unmarshal(message, &req); err != nil {
			controllerLog.Error("bad unmarshal message", slog.String("error", err.Error()))
			continue
		}

		handler, exists := h.customHandlers[req.Type]
		if !exists {
			controllerLog.Error("custom handler not found", slog.String("type", req.Type))
			return
		}
		handler.Handle(conn, &req)
	}
}

func SetConnectionTimeout(log *slog.Logger, conn net.Conn, timeout time.Duration) {
	if timeout > 0 {
		if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			log.Error("failed to set deadline", slog.String("error", err.Error()))
		}
	}
}

/*
if req.Token == "" {
        handlerlog.Error("missing token")
        _ = h.wr.WriteError(conn, "missing token") // use writer to send error
        return
    }
    // You need to inject the JWT secret into the handler. For simplicity, you can add it to Controller struct.
    // Let's add a field: jwtSecret string
    // Then validate:
    _, err = jwt.ValidateToken(req.Token, h.jwtSecret)
    if err != nil {
        handlerlog.Error("invalid token", slog.String("error", err.Error()))
        _ = h.wr.WriteError(conn, "invalid token")
        return
    }
*/
