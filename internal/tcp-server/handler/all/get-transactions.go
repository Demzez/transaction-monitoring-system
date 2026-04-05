package all

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type TransactionsGetter interface {
	GetTransactions() ([]dto.TransactionDTO, error)
}

type GetTransactionsHandler struct {
	log *slog.Logger
	db  TransactionsGetter
	wr  writers.WrInterface
}

func NewGetTransactionsHandler(log *slog.Logger, db TransactionsGetter, wr writers.WrInterface) *GetTransactionsHandler {
	return &GetTransactionsHandler{
		log: log,
		db:  db,
		wr:  wr,
	}
}

func (h *GetTransactionsHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.all.get-transactions.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	transactionDTOs, err := h.db.GetTransactions()
	if err != nil {
		handlerLog.Error("failed to get transaction", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	var protoAnswer []*protoStruct.RespTransaction
	for _, transaction := range transactionDTOs {
		protoAnswer = append(protoAnswer, transaction.DTOToProto())
	}

	data, err := proto.Marshal(&protoStruct.RespTransactionList{Transactions: protoAnswer})
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

func (h *GetTransactionsHandler) Type() string {
	return "get-transactions"
}
