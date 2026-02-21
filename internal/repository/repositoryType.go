package repository

import "time"

type Transaction struct {
	Hash        string
	Source      string
	Description string
	Type        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
