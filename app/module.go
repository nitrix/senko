package app

type ReplyFunc = func (response interface{}) error

type Module interface {
	OnRegister(store *Store)
	OnRequest(request interface{}, reply ReplyFunc) error
}
