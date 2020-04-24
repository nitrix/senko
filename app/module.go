package app

type Module interface {
	Dispatch(request Request, response Response) error
}
