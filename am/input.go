package am

import "io"

// Input represents a parsed and validated input
type Input map[string]interface{}

// InputService manages input lists for an organization and group.
type InputService interface {
	Init(config []byte) error
	Add(orgID, groupID, userID int64, rawInput io.Reader, input Input) error
	AddTo(orgID, groupID, userID, inputID int64, newInputs Input) error
	Get(orgID, groupID, userID, inputID int64) (Input, error)
	Update(orgID, groupID, userID int64, input Input) error
	Delete(orgID, groupID, userID, inputID int64) error
}

// InputReaderService read only implementation of above
type InputReaderService interface {
	Init(config []byte) error
	Get(orgID, groupID, userID, inputID int64) (Input, error)
}
