package store

import (
	"context"

	"gopkg.linkai.io/v1/repos/am/am"
)


// Storer for interfacing with the backend database for handling inputs
type Storer interface {
	Init(config []byte) error
	Create(ctx context.Context, newGroup *am.ScanGroup, newVersion *am.ScanGroupVersion) (oid int32, gid int32, err error)
	/*
		Add(orgID, groupID, userID int64, rawInput io.Reader, input am.Input) error
		AddTo(orgID, groupID, userID, inputID int64, newInputs am.Input) error
		Get(orgID, groupID, inputID int64) (am.Input, error)
		Update(orgID, groupID, userID int64, input am.Input) error
		Delete(orgID, groupID, userID, inputID int64) error
	*/
}
