package handler

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type TransactionGetter interface {
	GetTransaction(int64) (dto.TransactionDTO, error)
}

type GetTransactionHandler struct {
	log *slog.Logger
	db  TransactionGetter
	wr  writers.WrInterface
}

func NewGetTransactionHandler(log *slog.Logger, db TransactionGetter, wr writers.WrInterface) *GetTransactionHandler {
	return &GetTransactionHandler{
		log: log,
		db:  db,
		wr:  wr,
	}
}

func (h *GetTransactionHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.get-transaction.Process"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqTransaction
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("bad unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}
	if pd.TransactionId == 0 {
		handlerLog.Info("transaction_id is a required")
		if err := h.wr.WriteError(conn, "transaction_id is a required"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	transactionDTO, err := h.db.GetTransaction(pd.TransactionId)
	if err != nil {
		handlerLog.Error("failed to get transaction", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	data, err := proto.Marshal(transactionDTO.DTOToProto())
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

	handlerLog.Info("transaction successfully sent")
}

func (h *GetTransactionHandler) Type() string {
	return "get-transaction"
}
