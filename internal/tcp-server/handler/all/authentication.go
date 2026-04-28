package all

import (
	"log/slog"
	"net"
	"time"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Authenticator interface {
	AuthenticateUser(login, password string) (int64, error)
	GenerateNewUserToken(secret string, expiresIn time.Duration) (string, error)
}

type AuthenticationHandler struct {
	log           *slog.Logger
	service       Authenticator
	wr            writers.WrInterface
	jwtSecret     string
	tokenLifetime time.Duration
}

func NewAuthenticationHandler(log *slog.Logger, cfg *config.Config, service Authenticator, wr writers.WrInterface) *AuthenticationHandler {
	return &AuthenticationHandler{
		log:           log,
		service:       service,
		wr:            wr,
		jwtSecret:     cfg.JWT.Secret,
		tokenLifetime: cfg.JWT.ExpiryIn,
	}
}

func (h *AuthenticationHandler) Handle(conn net.Conn, req *protoStruct.Request) {

	const op = "internal.tcp-server.handler.all.authentication.Handle"

	handlerLog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)

	var pd protoStruct.ReqAuthentication
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerLog.Error("failed to unmarshal request", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	roleID, err := h.service.AuthenticateUser(pd.Login, pd.Password)
	if err != nil {
		if err = h.wr.WriteError(conn, "incorrect login or password"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	newToken, err := h.service.GenerateNewUserToken(h.jwtSecret, h.tokenLifetime)
	if err != nil {
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
	}

	protoAnswer := protoStruct.RespAuthentication{
		NewToken: newToken,
		RoleId:   roleID,
	}

	data, err := proto.Marshal(&protoAnswer)
	if err != nil {
		handlerLog.Error("failed to marshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "something went wrong"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	if err = h.wr.WriteResponse(conn, data); err != nil {
		handlerLog.Error("failed to response", slog.String("error", err.Error()))
	}

	handlerLog.Info("authentication succeed")
}

func (h *AuthenticationHandler) Type() string {
	return "authentication"
}
