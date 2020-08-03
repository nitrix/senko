package app

type Module interface {
	OnLoad(store *Store)
	OnUnload(store *Store)
	OnEvent(event *Event) error
}
