package save

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"transaction-monitoring-system/internal/repository"

	"github.com/go-playground/validator/v10"
)

type Request struct {
	Hash        string `json:"hash" validate:"required"`
	Source      string `json:"source" validate:"required"`
	Description string `json:"description" validate:"required"`
	Type        string `json:"type" validate:"required"`
	Status      string `json:"status" validate:"required"`
}

type TransactionSaver interface {
	SaveTransaction(transaction repository.TransactionDTO) error
}

func New(log *slog.Logger, saver TransactionSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.post-transaction.save.New"

		handlerlog := log.With(
			slog.String("op", op),
			slog.String("request_info", fmt.Sprintf("%s : %s : %s", r.Host, r.Method, r.URL.Path)),
		)

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			handlerlog.Error("failed to decode request body", slog.String("error", err.Error()))

			http.Error(w, "failed to decode request body", http.StatusBadRequest)

			return
		}
		if err = validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			handlerlog.Error("failed to validate request", slog.String("error", validateErr.Error()))

			http.Error(w, "failed to validate request", http.StatusBadRequest)

			return
		}

		err = saver.SaveTransaction(repository.TransactionDTO{
			Hash:        req.Hash,
			Source:      req.Source,
			Description: req.Description,
			Type:        req.Type,
			Status:      req.Status,
			CreatedAt:   time.Now(),
		})
		if err != nil {
			// TODO: add check for unique violation error, or doing something other
			//if errors.Is(err, repository.ErrTransactionAlreadyExists) {
			//	handlerlog.Error("transaction already exist", slog.String("error", err.Error()))
			//
			//	http.Error(w, "transaction already exist", http.StatusInternalServerError)
			//
			//	return
			//}
			handlerlog.Error("failed to save transaction", slog.String("error", err.Error()))

			http.Error(w, "failed to save transaction", http.StatusInternalServerError)

			return
		}

		handlerlog.Info("transaction successfully saved", slog.String("hash", req.Hash))
	}
}
