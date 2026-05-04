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

func BuildDoubtfulTransactionQuery(like string) (string, string) {
	base := `SELECT transaction_id, risk_score, description, decision FROM "doubtful_transaction"`
	
	if like == "" {
		return base, ""
	}
	
	q := base + `
		WHERE
			transaction_id::text LIKE $1 OR
			risk_score::text LIKE $1 OR
			description ILIKE $1 OR
			decision ILIKE $1`
	
	return q, "%" + like + "%"
}

func BuildFraudRuleQuery(like string) (string, string) {
	base := `SELECT rule_id, name, active, field_name, operator, value, add_risk FROM "fraud_rule"`
	
	if like == "" {
		return base, ""
	}
	
	q := base + `
		WHERE
			rule_id::text LIKE $1 OR
			name ILIKE $1 OR
			active::text LIKE $1 OR
			field_name ILIKE $1 OR
			operator ILIKE $1 OR
			value ILIKE $1 OR
			add_risk::text LIKE $1`
	
	return q, "%" + like + "%"
}

func BuildUserQuery(like string) (string, string) {
	base := `SELECT user_id, login, role_id, created_at FROM "user"`
	
	if like == "" {
		return base, ""
	}
	
	q := base + `
		WHERE
			user_id::text LIKE $1 OR
			login ILIKE $1 OR
			role_id::text LIKE $1`
	
	return q, "%" + like + "%"
}
