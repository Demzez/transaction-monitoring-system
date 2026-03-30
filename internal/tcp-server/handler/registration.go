package handler

import (
	"log/slog"
	"net"
	"time"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Registrator interface {
	Register(login string, password string, createdAt time.Time) error
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

func (h *RegistrationHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.registration.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqRegistration
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("bad unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
	}

	err := h.db.Register(pd.Login, pd.Password, time.Now())
	if err != nil {
		handlerLog.Error("failed to register", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, make([]byte, 0)); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("registration succeed")
}

func (h *RegistrationHandler) Type() string {
	return "registration"
}
