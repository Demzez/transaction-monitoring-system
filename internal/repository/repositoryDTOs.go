package repository

import "time"

type TransactionDTO struct {
	Hash        string
	Source      string
	Description string
	Type        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
