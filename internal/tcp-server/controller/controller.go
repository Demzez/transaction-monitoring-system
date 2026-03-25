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
	custom_error "transaction-monitoring-system/internal/lib/custom-error"
	"transaction-monitoring-system/protobuf"
	
	"google.golang.org/protobuf/proto"
)

type Handler interface {
	Handle(conn net.Conn, req *protobuf.Request)
	Type() string
}

type Controller struct { // TODO: передать сюда конфиг для проверки токена && refactoring!!!
	log         *slog.Logger
	handlers    map[string]Handler
	idleTimeout time.Duration
}

func NewController(log *slog.Logger, idleTimeout time.Duration, handlers ...Handler) *Controller {
	register := make(map[string]Handler)
	for _, h := range handlers {
		register[h.Type()] = h
	}
	
	return &Controller{
		log:         log,
		handlers:    register,
		idleTimeout: idleTimeout,
	}
}

func (c *Controller) Process(conn net.Conn, wg *sync.WaitGroup) {
	defer func() {
		conn.Close()
		wg.Done()
	}()
	
	const op = "internal.tcp-server.controller.handler.Process"
	
	controllerLog := c.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)
	
	controllerLog.Info("new client connected")
	
	for {
		err := setConnectionTimeout(controllerLog, conn, c.idleTimeout)
		if err != nil {
			return
		}
		
		length, err := readLengthPrefix(controllerLog, conn)
		if err != nil {
			return
		}
		
		message, err := readByteMessage(controllerLog, conn, length)
		if err != nil {
			return
		}
		
		req, err := byteToProtobufRequest(controllerLog, message)
		if err != nil {
			continue
		}
		
		handleConnection(controllerLog, conn, req, c.handlers)
	}
}

func setConnectionTimeout(log *slog.Logger, conn net.Conn, timeout time.Duration) error {
	if timeout >= 0 {
		if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			log.Error("failed to set timeout", slog.String("error", err.Error()))
			return custom_error.ErrFunc
		}
		
		return nil
	}
	
	log.Error("timeout is empty")
	return custom_error.ErrFunc
}

func readLengthPrefix(log *slog.Logger, conn net.Conn) (uint32, error) {
	const maxMessageLength = 4 << 20 // 4 MB
	
	lenBuf := make([]byte, 4)
	_, err := io.ReadFull(conn, lenBuf)
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Warn("timeout exceeded", slog.String("error", err.Error()))
			return 0, custom_error.ErrFunc
		}
		
		log.Error("something wrong with length prefix", slog.String("error", err.Error()))
		return 0, custom_error.ErrFunc
	}
	
	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		log.Info("request body is empty")
		return 0, custom_error.ErrFunc
	}
	if length > maxMessageLength {
		log.Info("request body is too long")
		return 0, custom_error.ErrFunc
	}
	
	return length, nil
}

func readByteMessage(log *slog.Logger, conn net.Conn, length uint32) ([]byte, error) {
	message := make([]byte, length)
	_, err := io.ReadFull(conn, message)
	if err != nil {
		log.Error("something wrong with payload", slog.String("error", err.Error()))
		return nil, custom_error.ErrFunc
	}
	
	return message, nil
}

func byteToProtobufRequest(log *slog.Logger, message []byte) (*protobuf.Request, error) {
	var req protobuf.Request
	if err := proto.Unmarshal(message, &req); err != nil {
		log.Error("bad unmarshal message", slog.String("error", err.Error()))
		return nil, custom_error.ErrFunc
	}
	
	return &req, nil
}

func handleConnection(log *slog.Logger, conn net.Conn, req *protobuf.Request, handlers map[string]Handler) {
	handler, exists := handlers[req.Type]
	if !exists {
		log.Error("handler not found", slog.String("type", req.Type))
		return
	}
	handler.Handle(conn, req)
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
