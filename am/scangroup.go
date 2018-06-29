package am

import (
	"context"
	"io"
)

// Input represents a parsed and validated input
type Input map[string]interface{}

// ScanGroupService manages input lists and configurations for an organization and group.
type ScanGroupService interface {
	Init(config []byte) error
	Add(ctx context.Context, orgID, groupID, userID int64, rawInput io.Reader, input Input) error
	AddTo(ctx context.Context, orgID, groupID, userID, inputID int64, newInputs Input) error
	Get(ctx context.Context, orgID, groupID, userID, inputID int64) (Input, error)
	Delete(ctx context.Context, orgID, groupID, userID, inputID int64) error
}

// ScanGroupReaderService read only implementation acquiring input lists and scan configs
type ScanGroupReaderService interface {
	Init(config []byte) error
	Get(ctx context.Context, orgID, groupID, userID int64) (Input, error)
}
