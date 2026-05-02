package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Creator interface {
	CreateFraudRule(rule dto.FraudRuleDTO) error
}

type CreateFraudRuleHandler struct {
	log     *slog.Logger
	service Creator
	wr      writers.WrInterface
}

func NewCreateFraudRuleHandler(log *slog.Logger, service Creator, wr writers.WrInterface) *CreateFraudRuleHandler {
	return &CreateFraudRuleHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *CreateFraudRuleHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.fraud.create-fraud-rule.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqChangeFraudRule
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("failed to unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
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

	err := h.service.CreateFraudRule(rule)
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, make([]byte, 0)); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("create fraud-rule succeed")
}

func (h *CreateFraudRuleHandler) Type() string {
	return "create-fraud-rule"
}
