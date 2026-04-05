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
	"transaction-monitoring-system/internal/lib/security/jwt"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Handler interface {
	Handle(conn net.Conn, req *protoStruct.Request)
	Type() string
}

type Controller struct {
	log         *slog.Logger
	handlers    map[string]Handler
	idleTimeout time.Duration
	jwtSecret   string
}

func NewController(log *slog.Logger, idleTimeout time.Duration, jwtSecret string, handlers ...Handler) *Controller {
	register := make(map[string]Handler)
	for _, h := range handlers {
		register[h.Type()] = h
	}

	return &Controller{
		log:         log,
		handlers:    register,
		idleTimeout: idleTimeout,
		jwtSecret:   jwtSecret,
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
			return
		}

		ok := validateTokenJWT(controllerLog, req, c.jwtSecret)
		if !ok {
			return
		}

		handleConnection(controllerLog, conn, req, c.handlers)
	}
}

func setConnectionTimeout(log *slog.Logger, conn net.Conn, timeout time.Duration) error {
	if timeout >= 0 {
		err := conn.SetDeadline(time.Now().Add(timeout))
		if err != nil {
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

		if errors.Is(err, io.EOF) {
			log.Warn("client disconnected", slog.String("error", err.Error()))
			return 0, custom_error.ErrFunc
		}

		log.Error("something wrong with length prefix", slog.String("error", err.Error()))
		return 0, custom_error.ErrFunc
	}

	length := binary.BigEndian.Uint32(lenBuf)
	if length == 0 {
		log.Warn("request body is empty")
		return 0, custom_error.ErrFunc
	}
	if length > maxMessageLength {
		log.Warn("request body is too long")
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

func byteToProtobufRequest(log *slog.Logger, message []byte) (*protoStruct.Request, error) {
	var req protoStruct.Request

	err := proto.Unmarshal(message, &req)
	if err != nil {
		log.Error("bad unmarshal message", slog.String("error", err.Error()))
		return nil, custom_error.ErrFunc
	}

	return &req, nil
}

func validateTokenJWT(log *slog.Logger, req *protoStruct.Request, jwtSecret string) bool {
	if req.Type == "authentication" || req.Type == "registration" {
		return true
	}

	if req.Token == "" {
		log.Error("missing token")
		return false
	}
	_, err := jwt.ValidateToken(req.Token, jwtSecret)
	if err != nil {
		log.Error("invalid token", slog.String("error", err.Error()))
		return false
	}

	return true
}

func handleConnection(log *slog.Logger, conn net.Conn, req *protoStruct.Request, handlers map[string]Handler) {
	handler, exists := handlers[req.Type]
	if !exists {
		log.Error("handler not found", slog.String("type", req.Type))
		return
	}
	handler.Handle(conn, req)
}
