package custom_handler

import (
	"log/slog"
	"net"
	"time"
	"transaction-monitoring-system/internal/tcp-server/writers"
	"transaction-monitoring-system/protobuf"
)

type AuthenticationHandler struct {
	log           *slog.Logger
	wr            writers.WrInterface
	jwtSecret     string
	tokenLifetime time.Duration
}

func NewAuthenticationHandler(log *slog.Logger, wr writers.WrInterface, jwtSecret string, tokenLifetime time.Duration) *AuthenticationHandler {
	return &AuthenticationHandler{
		log:           log,
		wr:            wr,
		jwtSecret:     jwtSecret,
		tokenLifetime: tokenLifetime,
	}
}

// TODO: написать эту чухню и можно садиться за клиента
func (h *AuthenticationHandler) Handle(conn net.Conn, req *protobuf.Request) {
	
}

func (h *AuthenticationHandler) Type() string {
	return "auth"
}
