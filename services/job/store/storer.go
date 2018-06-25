package store

import "gopkg.linkai.io/v1/repos/am/am"

// Storer is the interface for a backend data store such as postgres
type Storer interface {
	Init(config []byte) error
	Create(Job *am.Job)
	Start(orgID int64, jobID []byte) error
	Pause(orgID int64, jobID []byte) error
	Cancel(orgID int64, jobID []byte) error
}
