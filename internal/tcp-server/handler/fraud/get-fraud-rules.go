package fraud

import (
	"log/slog"
	"net"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type FraudRulesGetter interface {
	GetFraudRules() ([]dto.FraudRuleDTO, error)
}

type GetFraudRulesHandler struct {
	log     *slog.Logger
	service FraudRulesGetter
	wr      writers.WrInterface
}

func NewGetFraudRulesHandler(log *slog.Logger, service FraudRulesGetter, wr writers.WrInterface) *GetFraudRulesHandler {
	return &GetFraudRulesHandler{
		log:     log,
		service: service,
		wr:      wr,
	}
}

func (h *GetFraudRulesHandler) Handle(conn net.Conn, req *protoStruct.Request) {
	const op = "internal.tcp-server.handler.fraud.get-fraud-rules.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	fraudRulesDTOs, err := h.service.GetFraudRules()
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	var protoAnswer []*protoStruct.RespFraudRule
	for _, fraudRule := range fraudRulesDTOs {
		protoAnswer = append(protoAnswer, fraudRule.DTOToProto())
	}

	data, err := proto.Marshal(&protoStruct.RespFraudRuleList{FraudRules: protoAnswer})
	if err != nil {
		handlerLog.Error("failed to marshal fraud-rules", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, data); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("fraud-rules successfully sent")
}

func (h *GetFraudRulesHandler) Type() string {
	return "get-fraud-rules"
}
