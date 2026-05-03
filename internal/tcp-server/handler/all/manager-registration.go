package all

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"
	
	"google.golang.org/protobuf/proto"
)

type Registrator interface {
	RegisterManager(login string, password string) error
}

type ManagerRegistrationHandler struct {
	log     *slog.Logger
	service Registrator
	wr      writers.WrInterface
}

func NewManagerRegistrationHandler(log *slog.Logger, service Registrator, wr writers.WrInterface) *ManagerRegistrationHandler {
	return &ManagerRegistrationHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *ManagerRegistrationHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	
	const op = "internal.tcp-server.handler.all.manager-registration.Handle"
	
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
	
	err := h.service.RegisterManager(pd.Login, pd.Password)
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}
	
	data := make([]byte, 0)
	if err = h.wr.WriteResponse(conn, data); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}
	
	handlerLog.Info("manager registration succeed")
}

func (h *ManagerRegistrationHandler) Type() string {
	return "manager-registration"
}
