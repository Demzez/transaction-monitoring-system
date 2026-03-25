package handler

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protobuf"
)

type Registrator interface {
	Register(dto dto.UserDTO) error
}

type RegistrationHandler struct {
	log *slog.Logger
	db  Registrator
	wr  writers.WrInterface
}

func NewRegistrationHandler(log *slog.Logger, db Registrator, wr writers.WrInterface) *RegistrationHandler {
	return &RegistrationHandler{
		log: log,
		db:  db,
		wr:  wr,
	}
}

func (h *RegistrationHandler) Handle(conn net.Conn, req *protobuf.Request) {

	const op = "internal.tcp-server.handler.registration.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	handlerLog.Info("")
}

func (h *RegistrationHandler) Type() string {
	return "registration"
}
