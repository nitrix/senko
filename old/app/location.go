package app

type Location struct {
	Code string
	IsInUse func() bool
}
