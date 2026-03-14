package postgres

import (
	"testing"
	"time"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestStorage_SaveTransaction(t *testing.T) {
	cases := []struct {
		name        string
		transaction dto.TransactionDTO
		respError   string
	}{
		{
			name: "Success",
			transaction: dto.TransactionDTO{
				Hash:        "rtgrbe7rew343rnjuh893h",
				Source:      "localhost",
				Description: "test transaction",
				Type:        "firstType",
				Status:      "testStatus",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		{
			name: "Error unique hash rule",
			transaction: dto.TransactionDTO{
				Hash:        "rtgrbe7rew343rnjuh893h",
				Source:      "localhost",
				Description: "test transaction",
				Type:        "firstType",
				Status:      "testStatus",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			respError: "internal.repository.postgres.SaveTransaction : " + repository.ErrTransactionAlreadyExists.Error(),
		},
	}
	cfg := config.MustLoad()
	storage, err := New(cfg.PostgresDB)
	if err != nil {
		t.Error("failed to initialize postgres storage")
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err = storage.SaveTransaction(tc.transaction)
			if err == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.respError)
			}
		})
	}
}

func TestStorage_DeleteTransaction(t *testing.T) {
	cases := []struct {
		name            string
		transactionHash string
		respError       string
	}{
		{
			name:            "Success",
			transactionHash: "rtgrbe7rew343rnjuh893h",
		},
		{
			name:            "Error transaction not found",
			transactionHash: "rtgrbe7rew343rnjuh893h",
			respError:       "internal.repository.postgres.DeleteTransaction : " + repository.ErrTransactionNotFound.Error(),
		},
	}
	cfg := config.MustLoad()
	storage, err := New(cfg.PostgresDB)
	if err != nil {
		t.Error("failed to initialize postgres storage")
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			err = storage.DeleteTransaction(tc.transactionHash)
			if err == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.respError)
			}
		})
	}
}
