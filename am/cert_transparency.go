package am

import "context"

// CTServer holds the state of a certificate transparency server
type CTServer struct {
	ID              int    `json:"id"`
	URL             string `json:"url"`
	Index           int64  `json:"index"`
	IndexUpdated    int64  `json:"index_updated"`
	Step            int    `json:"step"`
	TreeSize        int64  `json:"tree_size"`
	TreeSizeUpdated int64  `json:"tree_size_updated"`
}

type CTCoordinatorStatus int32

const (
	Unknown CTCoordinatorStatus = 0
	Stopped CTCoordinatorStatus = 1
	Started CTCoordinatorStatus = 2
)

const (
	CTCoordinatorServiceKey = "ctcoordinatorservice"
	CTWorkerServiceKey      = "ctworkerservice"
)

type CTCoordinatorService interface {
	GetStatus(ctx context.Context) (CTCoordinatorStatus, int32, error)
	SetStatus(ctx context.Context, status CTCoordinatorStatus) (int32, error)
	AddWorker(ctx context.Context, workerCount int32) error
	RemoveWorker(ctx context.Context, workerCount int32) error
	UpdateDuration(ctx context.Context, newDuration int64) error
}

type CTWorkerService interface {
	GetCTCertificates(ctx context.Context, ctServer *CTServer) (*CTServer, error)
	SetExtractors(ctx context.Context, numExtractors int32) error
}
