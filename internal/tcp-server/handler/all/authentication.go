package all

import (
	"log/slog"
	"net"
	"time"
	"transaction-monitoring-system/internal/lib/security/jwt"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protoStruct"

	"google.golang.org/protobuf/proto"
)

type Authenticator interface {
	Authenticate(username, password string) error
}

type AuthenticationHandler struct {
	log           *slog.Logger
	db            Authenticator
	wr            writers.WrInterface
	jwtSecret     string
	tokenLifetime time.Duration
}

func NewAuthenticationHandler(log *slog.Logger, db Authenticator, wr writers.WrInterface, jwtSecret string, tokenLifetime time.Duration) *AuthenticationHandler {
	return &AuthenticationHandler{
		log:           log,
		db:            db,
		wr:            wr,
		jwtSecret:     jwtSecret,
		tokenLifetime: tokenLifetime,
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
		handlerLog.Error("bad unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerLog.Error("failed to response with error", slog.String("error", err.Error()))
		}
		return
	}

	err := h.db.Authenticate(pd.Login, pd.Password)
	if err != nil {
		handlerLog.Error("failed to authenticate", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "incorrect login or password"); err != nil {
			handlerLog.Error("failed to write response with error", slog.String("error", err.Error()))
		}
		return
	}

	newToken, _ := jwt.GenerateToken(h.jwtSecret, h.tokenLifetime)
	protoAnswer := protoStruct.RespAuthentication{
		NewToken: newToken,
	}

	data, err := proto.Marshal(&protoAnswer)
	if err != nil {
		handlerLog.Error("failed to marshal token", slog.String("error", err.Error()))
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
