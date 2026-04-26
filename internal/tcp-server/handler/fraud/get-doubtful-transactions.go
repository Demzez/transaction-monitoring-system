package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type DoubtfulTransactionsGetter interface {
	GetDoubtfulTransactions() ([]dto.DoubtfulTransactionDTO, error)
}

type GetDoubtfulTransactionsHandler struct {
	log     *slog.Logger
	service DoubtfulTransactionsGetter
	wr      writers.WrInterface
}

func NewGetDoubtfulTransactionsHandler(log *slog.Logger, service DoubtfulTransactionsGetter, wr writers.WrInterface) *GetDoubtfulTransactionsHandler {
	return &GetDoubtfulTransactionsHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *GetDoubtfulTransactionsHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.fraud.get-doubtful-transactions.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	dlTransactionDTOs, err := h.service.GetDoubtfulTransactions()
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	var protoAnswer []*protoStruct.RespDoubtfulTransaction
	for _, dlTransaction := range dlTransactionDTOs {
		protoAnswer = append(protoAnswer, dlTransaction.DTOToProto())
	}

	data, err := proto.Marshal(&protoStruct.RespDoubtfulTransactionList{DoubtfulTransactions: protoAnswer})
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

	handlerLog.Info("doubtful-transactions successfully sent")
}

func (h *GetDoubtfulTransactionsHandler) Type() string {
	return "get-doubtful-transactions"
}
