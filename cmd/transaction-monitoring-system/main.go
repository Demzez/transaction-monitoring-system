package main //TODO: как то рефакторить main

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
	"transaction-monitoring-system/internal/http-server/handler/receive"
	"transaction-monitoring-system/internal/lib/logger/slog/slogpretty"
	"transaction-monitoring-system/internal/repository/postgres"
	transaction_service "transaction-monitoring-system/internal/service/transaction-service"
	user_service "transaction-monitoring-system/internal/service/user-service"
	"transaction-monitoring-system/internal/tcp-server/controller"
	"transaction-monitoring-system/internal/tcp-server/handler/admin"
	"transaction-monitoring-system/internal/tcp-server/handler/all"
	"transaction-monitoring-system/internal/tcp-server/handler/fraud"
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
	log.Info("database is connected", slog.String("connection_pool", repository.Statistic()))

	//init services
	transactionService := transaction_service.NewTransactionService(log, repository)
	userService := user_service.NewUserService(log, repository)

	// init http && tcp servers
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = newHttpServer(sigContext, log, cfg, transactionService)
		if err != nil {
			log.Error("failed in initHTTPserver", slog.String("error", err.Error()))
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = newTCPServer(sigContext, log, cfg, transactionService, userService)
		if err != nil {
			log.Error("failed in initTCPserver", slog.String("error", err.Error()))
			return
		}
	}()

	wg.Wait()
	repository.Close()
	log.Info("----------------------graceful shutdown is completed----------------------")
}

func newHttpServer(sigCtx context.Context, log *slog.Logger, cfg *config.Config, tService *transaction_service.TransactionService) error {
	muxRouter := http.NewServeMux()
	muxRouter.HandleFunc("POST /send", receive.New(tService))

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

func newTCPServer(sigCtx context.Context, log *slog.Logger, cfg *config.Config, tService *transaction_service.TransactionService, uService *user_service.UserService) error {
	wr := &writers.ProtobufWriter{}
	newController := controller.NewController(log, cfg,
		all.NewManagerRegistrationHandler(log, uService, wr),
		all.NewAuthenticationHandler(log, cfg, uService, wr),
		all.NewGetTransactionHandler(log, tService, wr),
		all.NewGetTransactionsHandler(log, tService, wr),
		fraud.NewGetDoubtfulTransactionsHandler(log, tService, wr),
		fraud.NewChangeDecisionHandler(log, tService, wr),
		fraud.NewGetFraudRulesHandler(log, tService, wr),
		fraud.NewChangeFraudRuleHandler(log, tService, wr),
		fraud.NewCreateFraudRuleHandler(log, tService, wr),
		fraud.NewDeleteFraudRuleHandler(log, tService, wr),
		admin.NewFraudRegistrationHandler(log, uService, wr),
		admin.NewAdminRegistrationHandler(log, uService, wr),
		admin.NewGetUsersHandler(log, uService, wr),
		admin.NewDeleteUserHandler(log, uService, wr),
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
