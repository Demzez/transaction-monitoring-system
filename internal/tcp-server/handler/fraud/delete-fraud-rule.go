package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Deleter interface {
	DeleteFraudRule(ruleId int64) error
}

type DeleteFraudRuleHandler struct {
	log     *slog.Logger
	service Deleter
	wr      writers.WrInterface
}

func NewDeleteFraudRuleHandler(log *slog.Logger, service Deleter, wr writers.WrInterface) *DeleteFraudRuleHandler {
	return &DeleteFraudRuleHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *DeleteFraudRuleHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.fraud.delete-fraud-rule.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqFraudRule
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("failed to unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	err := h.service.DeleteFraudRule(pd.RuleId)
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

	handlerLog.Info("fraud-rule successfully deleted")
}

func (h *DeleteFraudRuleHandler) Type() string { return "delete-fraud-rule" }
