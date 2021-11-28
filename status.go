package runner

type status int

const (
	ready     status = iota
	running   status = iota
	completed status = iota
	stopped   status = iota
	failed    status = iota
)
