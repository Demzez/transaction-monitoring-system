package admin

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"
	
	"google.golang.org/protobuf/proto"
)

type Deleter interface {
	DeleteUser(userId int64) error
}

type DeleteUserHandler struct {
	log     *slog.Logger
	service Deleter
	wr      writers.WrInterface
}

func NewDeleteUserHandler(log *slog.Logger, service Deleter, wr writers.WrInterface) *DeleteUserHandler {
	return &DeleteUserHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *DeleteUserHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.admin.delete-user.Handle"
	
	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)
	
	var pd protoStruct.ReqUser
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("failed to unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}
	
	err := h.service.DeleteUser(pd.UserId)
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
	
	handlerLog.Info("user successfully deleted")
}

func (h *DeleteUserHandler) Type() string { return "delete-user" }
