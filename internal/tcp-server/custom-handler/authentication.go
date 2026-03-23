package custom_handler

import (
	"log/slog"
	"net"
	"time"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protobuf"
	
	"google.golang.org/protobuf/proto"
)

type Authenticator interface {
	Authenticate(username, password string) (string, error)
}

type AuthenticationHandler struct {
	log           *slog.Logger
	repository    Authenticator
	wr            writers.WrInterface
	jwtSecret     string
	tokenLifetime time.Duration
}

func NewAuthenticationHandler(log *slog.Logger, db Authenticator, wr writers.WrInterface, jwtSecret string, tokenLifetime time.Duration) *AuthenticationHandler {
	return &AuthenticationHandler{
		log:           log,
		wr:            wr,
		jwtSecret:     jwtSecret,
		tokenLifetime: tokenLifetime,
	}
}

func (h *AuthenticationHandler) Handle(conn net.Conn, req *protobuf.Request) {
	
	const op = "internal.tcp-server.custom-handler.authentication.Handle"
	
	handlerlog := h.log.With(
		slog.String("op", op),
		slog.String("remoteAddr", conn.RemoteAddr().String()),
	)
	
	var pd protobuf.AuthenticationRequest
	if err := proto.Unmarshal(req.Payload, &pd); err != nil {
		handlerlog.Error("bad unmarshal payload", slog.String("error", err.Error()))
		if err = h.wr.WriteError(conn, "bad request"); err != nil {
			handlerlog.Error("failed to response with error", slog.String("error", err.Error()))
		}
	}
	
	// TODO: metod for authorization login and password in db
}

func (h *AuthenticationHandler) Type() string {
	return "auth"
}
