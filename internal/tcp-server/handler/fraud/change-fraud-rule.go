package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type FraudRuleChanger interface {
	ChangeFraudRule(rule dto.FraudRuleDTO) error
}

type ChangeFraudRuleHandler struct {
	log     *slog.Logger
	service FraudRuleChanger
	wr      writers.WrInterface
}

func NewChangeFraudRuleHandler(log *slog.Logger, service FraudRuleChanger, wr writers.WrInterface) *ChangeFraudRuleHandler {
	return &ChangeFraudRuleHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *ChangeFraudRuleHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.fraud.change-fraud-rules.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqChangeFraudRule
	err := proto.Unmarshal(req.Payload, &pd)
	if err != nil {
		handlerLog.Error("failed to unmarshal request", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
	}

	rule := dto.FraudRuleDTO{
		RuleID:    pd.RuleId,
		Name:      pd.Name,
		Active:    pd.Active,
		FieldName: pd.FieldName,
		Operator:  pd.Operator,
		Value:     pd.Value,
		AddRisk:   pd.AddRisk,
	}

	err = h.service.ChangeFraudRule(rule)
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

	handlerLog.Info("fraud-rule successfully changed")
}

func (h *ChangeFraudRuleHandler) Type() string {
	return "change-fraud-rules"
}
