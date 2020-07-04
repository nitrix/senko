package app

type Module interface {
	Load() error
	Unload() error

	OnCommand(event *CommandEvent) error
	OnMessageCreated(event *MessageCreatedEvent) error
}
