package app

type Gateway interface {
	Name() string
	OnRegister()

	Run(app *App) error
	Stop()
}
