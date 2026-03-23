package custom_handler

type Registrator interface {
	Register(username, password string) error
}

// TODO: releas registration, and after that you must to authentication
