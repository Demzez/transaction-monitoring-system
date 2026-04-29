package admin

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type UsersGetter interface {
	GetUsers() ([]dto.UserDTO, error)
}

type GetUsersHandler struct {
	log     *slog.Logger
	service UsersGetter
	wr      writers.WrInterface
}

func NewGetUsersHandler(log *slog.Logger, service UsersGetter, wr writers.WrInterface) *GetUsersHandler {
	return &GetUsersHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *GetUsersHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.admin.get-users.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	userDTOs, err := h.service.GetUsers()
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	var protoAnswer []*protoStruct.RespUser
	for _, user := range userDTOs {
		protoAnswer = append(protoAnswer, user.DTOToProto())
	}

	data, err := proto.Marshal(&protoStruct.RespUserList{Users: protoAnswer})
	if err != nil {
		handlerLog.Error("failed to marshal transaction", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, data); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("users successfully sent")
}

func (h *GetUsersHandler) Type() string {
	return "get-users"
}
