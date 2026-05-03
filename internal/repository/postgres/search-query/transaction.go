package search_query

func BuildTransactionQuery(like string) (string, string) {

	base := `SELECT transaction_id, hash, source, amount, direction, status, created_at, updated_at FROM "transaction"`

	if like == "" {
		return base, ""
	}

	q := base + `
		WHERE
			hash ILIKE $1 OR
			source ILIKE $1 OR
			amount::text LIKE $1 OR
			direction ILIKE $1 OR
			status ILIKE $1`

	return q, "%" + like + "%"
}
