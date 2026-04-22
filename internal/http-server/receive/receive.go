package receive //TODO: придумать другое название, скорее всего это будет handler, дальше service(в котором будет лежать логика антифрода), ну и репозиторий

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"transaction-monitoring-system/internal/dto"

	"github.com/go-playground/validator/v10"
)

type Request struct {
	Hash      string `json:"hash" validate:"required"`
	Source    string `json:"source" validate:"required"`
	Amount    int64  `json:"amount" validate:"required"`
	Direction string `json:"direction" validate:"required"`
	Status    string `json:"status" validate:"required"`
}

type FraudService interface {
	Control(transaction dto.TransactionDTO) error
}

func New(log *slog.Logger, fService FraudService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "internal.http-server.receive.save.New"

		var req Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "failed to decode request body", http.StatusBadRequest)

			return
		}
		if err = validator.New().Struct(req); err != nil {
			http.Error(w, "failed to validate request", http.StatusBadRequest)

			return
		}

		transaction := dto.TransactionDTO{
			Hash:      req.Hash,
			Source:    req.Source,
			Amount:    req.Amount,
			Direction: req.Direction,
			Status:    req.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err = fService.Control(transaction)
		if err != nil {
			http.Error(w, "failed to save transaction", http.StatusInternalServerError)

			return
		}
	}
}
