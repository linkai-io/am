package state

import "gopkg.linkai.io/v1/repos/am/am"

// Stater is for interfacing with a state management system
// It is responsible for managing the life cycle of a job
type Stater interface {
	Init(config []byte) error
	Create(job *am.Job) error
	Finalize(job *am.Job) error
	Populate()
}
