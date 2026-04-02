package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/http-server/post-transaction/save"
	"transaction-monitoring-system/internal/lib/logger/slog/slogpretty"
	"transaction-monitoring-system/internal/repository/postgres"
	"transaction-monitoring-system/internal/tcp-server/controller"
	"transaction-monitoring-system/internal/tcp-server/handler"
	"transaction-monitoring-system/internal/tcp-server/writers"
)

func main() {
	// init config
	cfg := config.MustLoad()

	// init logger
	log := slogpretty.SetupPrettyLogger()
	log.Info("config read")

	// graceful shutdown context
	sigContext, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	// start servers
	wg := &sync.WaitGroup{}
	// init http router & server
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = initHttpServer(sigContext, cfg, log, repository)
		if err != nil {
			log.Error("failed in initHTTPserver", slog.String("error", err.Error()))
			return
		}
	}()

	// init tcp server
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = initTCPServer(sigContext, cfg, log, repository)
		if err != nil {
			log.Error("failed in initTCPserver", slog.String("error", err.Error()))
			return
		}
	}()

	wg.Wait()
	repository.Close()
	log.Info("----------------------graceful shutdown is completed----------------------")
}

func initHttpServer(sigCtx context.Context, cfg *config.Config, log *slog.Logger, repository *postgres.Repository) error {
	muxRouter := http.NewServeMux()
	muxRouter.HandleFunc("POST /post-transaction", save.New(log, repository))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      muxRouter,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	log.Info("HTTP server is configured")

	go func() {
		<-sigCtx.Done()
		log.Info("HTTP server received shutdown signal, processing shutdown...")

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error("HTTP server shutdown error", slog.String("error", err.Error()))
		}
	}()

	log.Info("---HTTP SERVER START---", slog.String("http://address/...", cfg.HTTPServer.Address))
	acErr := srv.ListenAndServe()
	if acErr != nil {
		if errors.Is(acErr, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("failed to start HTTP server")
	}
	return nil
}

func initTCPServer(sigCtx context.Context, cfg *config.Config, log *slog.Logger, repository *postgres.Repository) error {
	wr := &writers.ProtobufWriter{}
	newController := controller.NewController(log, cfg.TCPServer.IdleTimeout, cfg.JWT.Secret,
		handler.NewRegistrationHandler(log, repository, wr),
		handler.NewAuthenticationHandler(log, repository, wr, cfg.JWT.Secret, cfg.JWT.ExpiryIn),
		handler.NewGetTransactionsHandler(log, repository, wr),
	)

	listener, err := net.Listen("tcp", cfg.TCPServer.Address)
	if err != nil {
		return fmt.Errorf("failed to config TCP server")
	}
	log.Info("TCP server is configured")

	go func() {
		<-sigCtx.Done()
		log.Info("TCP server received shutdown signal, processing shutdown...")

		if err = listener.Close(); err != nil {
			log.Error("TCP server shutdown error", slog.String("error", err.Error()))
		}
	}()

	log.Info("---TCP SERVER START---", slog.String("address", cfg.TCPServer.Address))
	wgClient := &sync.WaitGroup{}
	for {
		conn, acErr := listener.Accept()
		if acErr != nil {
			if errors.Is(acErr, net.ErrClosed) {
				log.Info("TCP server is closed, wait for all clients to finish...")
				wgClient.Wait()
				return nil
			}
			log.Error("failed to accept connection", slog.String("error", err.Error()))
			continue
		}

		wgClient.Add(1)
		go newController.Process(conn, wgClient)
	}
}
