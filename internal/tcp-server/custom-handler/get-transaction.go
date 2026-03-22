package custom_handler

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protobuf"

	"google.golang.org/protobuf/proto"
)

type transactionGetter interface {
	GetTransaction(int64, *dto.TransactionDTO) error
}

type GetTransactionHandler struct {
	log *slog.Logger
	db  transactionGetter
	wr  writers.WrInterface
}

func NewGetTransactionHandler(log *slog.Logger, db transactionGetter, wr writers.WrInterface) *GetTransactionHandler {
	return &GetTransactionHandler{
		log: log,
		db:  db,
		wr:  wr,
	}
}

func (h *GetTransactionHandler) Handle(conn net.Conn, req *protobuf.Request) {

	const op = "internal.tcp-server.custom-handler.GetTransactionHandler.Handle"

	handlerlog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protobuf.PullTransaction
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerlog.Error("bad unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad unmarshal payload"); err != nil {
			handlerlog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}
	if pd.TransactionId == 0 {
		handlerlog.Info("transaction_id is a required")
		if err := h.wr.WriteError(conn, "transaction_id is a required"); err != nil {
			handlerlog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	var transactionDTO dto.TransactionDTO
	err := h.db.GetTransaction(pd.TransactionId, &transactionDTO)
	if err != nil {
		handlerlog.Error("failed to get transaction", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "failed to get transaction"); err != nil {
			handlerlog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	data, err := proto.Marshal(transactionDTO.DTOToProto())
	if err != nil {
		handlerlog.Error("failed to marshal transaction", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "failed to marshal transaction"); err != nil {
			handlerlog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, data); err != nil {
		handlerlog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerlog.Info("transaction successfully sent")
}

func (h *GetTransactionHandler) Type() string {
	return "get-transaction"
}
