package postgres

import (
	"fmt"
	"testing"
	"time"
	"transaction-monitoring-system/internal/config"
	"transaction-monitoring-system/internal/dto"
	"transaction-monitoring-system/internal/repository"

	"github.com/stretchr/testify/require"
)

func TestRepository_SaveTransaction(t *testing.T) {
	cases := []struct {
		name        string
		transaction dto.TransactionDTO
		respError   string
	}{
		{
			name: "Success",
			transaction: dto.TransactionDTO{
				Hash:      "rtgrbe7rew343rnjuh893h",
				Source:    "localhost",
				Amount:    3000,
				Direction: "firstD",
				Status:    "testStatus",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		{
			name: "Error unique hash rule",
			transaction: dto.TransactionDTO{
				Hash:      "rtgrbe7rew343rnjuh893h",
				Source:    "localhost",
				Amount:    3000,
				Direction: "firstD",
				Status:    "testStatus",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			respError: "internal.repository.postgres.transaction.SaveTransaction : " + repository.ErrRecordAlreadyExists.Error(),
		},
	}
	cfg := config.MustLoad()
	repo, err := New(cfg.PostgresDB)
	if err != nil {
		t.Error("failed to initialize postgres repo")
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			id, err := repo.SaveTransaction(tc.transaction)
			fmt.Println("transaction id: ", id)
			if err == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.respError)
			}
		})
	}
}

func TestRepository_DeleteTransaction(t *testing.T) {
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
			respError:       "internal.repository.postgres.transaction.DeleteTransaction : " + repository.ErrRecordNotFound.Error(),
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
			err = storage.DeleteTransactionByHash(tc.transactionHash)
			if err == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.respError)
			}
		})
	}
}
