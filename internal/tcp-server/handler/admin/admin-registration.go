package admin

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Registrator interface {
	RegisterFraudSpecialist(login string, password string) error
	RegisterAdmin(login string, password string) error
}

type AdminRegistrationHandler struct {
	log     *slog.Logger
	service Registrator
	wr      writers.WrInterface
}

func NewAdminRegistrationHandler(log *slog.Logger, service Registrator, wr writers.WrInterface) *AdminRegistrationHandler {
	return &AdminRegistrationHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *AdminRegistrationHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.admin.admin-registration.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqRegistration
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("failed to unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
	}

	err := h.service.RegisterAdmin(pd.Login, pd.Password)
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, make([]byte, 0)); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("admin registration succeed")
}

func (h *AdminRegistrationHandler) Type() string {
	return "admin-registration"
}
