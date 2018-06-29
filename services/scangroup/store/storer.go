package store

// Storer for interfacing with the backend database for handling inputs
type Storer interface {
	Init(config []byte) error
	/*
		Add(orgID, groupID, userID int64, rawInput io.Reader, input am.Input) error
		AddTo(orgID, groupID, userID, inputID int64, newInputs am.Input) error
		Get(orgID, groupID, inputID int64) (am.Input, error)
		Update(orgID, groupID, userID int64, input am.Input) error
		Delete(orgID, groupID, userID, inputID int64) error
	*/
}
