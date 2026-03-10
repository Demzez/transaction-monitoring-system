package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/http-server/post-transaction/save"
	"transaction-monitoring-system/internal/lib/logger/slog/slogpretty"
	"transaction-monitoring-system/internal/repository/postgres"
	socket_connection "transaction-monitoring-system/internal/tcp-server/socket-connection"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := slogpretty.SetupPrettyLogger()
	log.Info("config read")

	// init database
	repository, err := postgres.New(cfg.PostgresDB)
	if err != nil {
		log.Error("database is not initialized", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if repository == nil {
		log.Error("database is not initialized correctly")
		os.Exit(1)
	}
	log.Info("database is connected", slog.String("connection_pool", repository.Statistic()))

	// init http router & server
	go initHttpServer(cfg, log, repository)

	// init tcp server
	initTCPServer(cfg, log, repository)
}

func initHttpServer(cfg *config.Config, log *slog.Logger, repository *postgres.Repository) {
	muxRouter := http.NewServeMux()
	muxRouter.HandleFunc("POST /post-transaction", save.New(log, repository))
	log.Info("router for HTTP is initialized")

	log.Info("---HTTP SERVER START---", slog.String("http://address/...", cfg.HTTPServer.Address))
	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      muxRouter,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if srv.ListenAndServe() != nil {
		log.Error("failed to start HTTP server")
		os.Exit(1)
	}
}

func initTCPServer(cfg *config.Config, log *slog.Logger, repository *postgres.Repository) {
	log.Info("---TCP SERVER START---", slog.String("address", cfg.HTTPServer.Address))
	listener, err := net.Listen("tcp", cfg.TCPServer.Address)
	if err != nil {
		log.Error("failed to start TCP server")
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error("failed to accept connection", slog.String("error", err.Error()))
			continue
		}
		go socket_connection.NewHandler(log, repository).Handle(conn)
	}
}
