package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type DecisionChanger interface {
	ChangeDecision(transactionId int64, decision string) error
}

type ChangeDecisionHandler struct {
	log     *slog.Logger
	service DecisionChanger
	wr      writers.WrInterface
}

func NewChangeDecisionHandler(log *slog.Logger, service DecisionChanger, wr writers.WrInterface) *ChangeDecisionHandler {
	return &ChangeDecisionHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *ChangeDecisionHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.fraud.change-decison.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqChangeDecision
	err := proto.Unmarshal(req.Payload, &pd)
	if err != nil {
		handlerLog.Error("failed to unmarshal request", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
	}

	err = h.service.ChangeDecision(pd.TransactionId, pd.Decision)
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

	handlerLog.Info("decision successfully changed")
}

func (h *ChangeDecisionHandler) Type() string {
	return "change-decision"
}
