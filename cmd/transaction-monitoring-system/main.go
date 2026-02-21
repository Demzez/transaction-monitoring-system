package main

import (
	"log/slog"
	"os"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/lib/logger/slog/slogpretty"
	"transaction-monitoring-system/internal/repository/postgres"
)

func main() {
	// TODO: init config
	cfg := config.MustLoad()

	// TODO: init logger
	log := setupPrettyLogger()
	log.Info("config read")

	// TODO: init database
	storage, err := postgres.New(cfg.PostgresDB)
	if err != nil {
		log.Error("database is not initialized", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if storage == nil {
		log.Warn("database is not initialized correctly")
		os.Exit(1)
	}
	log.Info("database is connected", slog.String("connection_pool", storage.Statistic()))

	//err = storage.SaveTransaction(repository.Transaction{
	//	Hash:        "something",
	//	Source:      "transaction-monitoring-system",
	//	Description: "something",
	//	Type:        "transaction-monitoring-system",
	//	Status:      "success",
	//	CreatedAt:   time.Now(),
	//	UpdatedAt:   time.Now(),
	//})
	//if err != nil {
	//	log.Error("save is not success", slog.String("error", err.Error()))
	//}
	//
	err = storage.DeleteTransaction("rtgrbe7rew343rnjuh893h")
	if err != nil {
		log.Error("delete is not success", slog.String("error", err.Error()))
	}

	// TODO: init router

	// TODO: init server
}

func setupPrettyLogger() *slog.Logger {

	loggerOptions := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}

	return slog.New(loggerOptions.NewPrettyHandler(os.Stdout))
}
