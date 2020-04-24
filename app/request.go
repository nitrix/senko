package app

type Request struct {
	Args []string // FIXME: private with convenient accessors
	NSFW bool
}
