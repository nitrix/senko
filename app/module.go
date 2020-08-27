package app

type Module interface {
	OnRegister(store *Store)
	OnEvent(gateway *Gateway, event interface{}) error
}