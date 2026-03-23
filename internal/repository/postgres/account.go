package postgres

func (r *Repository) Register(username, password string) error {
	return nil
}

func (r *Repository) Authenticate(username, password string) (string, error) {
	return "", nil
}
