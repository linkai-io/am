package protoc

import (
	"context"

	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/protocservices/certtranscoordinator"
	"github.com/rs/zerolog/log"
)

type CTCoordProtocService struct {
	cs am.CTCoordinatorService
}

func New(implementation am.CTCoordinatorService) *CTCoordProtocService {
	return &CTCoordProtocService{cs: implementation}
}

func (c *CTCoordProtocService) GetStatus(ctx context.Context, in *certtranscoordinator.GetStatusRequest) (*certtranscoordinator.StatusResponse, error) {
	status, workerCount, err := c.cs.GetStatus(ctx)
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct coordinator get status")
		return nil, err
	}

	return &certtranscoordinator.StatusResponse{Status: certtranscoordinator.CTStatus(status), WorkerCount: workerCount}, nil
}

func (c *CTCoordProtocService) SetStatus(ctx context.Context, in *certtranscoordinator.SetStatusRequest) (*certtranscoordinator.StatusResponse, error) {
	workerCount, err := c.cs.SetStatus(ctx, am.CTCoordinatorStatus(in.Status))
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct coordinator set status")
		return nil, err
	}

	return &certtranscoordinator.StatusResponse{Status: in.Status, WorkerCount: workerCount}, nil
}

func (c *CTCoordProtocService) AddWorker(ctx context.Context, in *certtranscoordinator.AddWorkerRequest) (*certtranscoordinator.WorkerAddedResponse, error) {
	err := c.cs.AddWorker(ctx, in.WorkerCount)
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct coordinator add worker")
		return nil, err
	}

	return &certtranscoordinator.WorkerAddedResponse{}, nil
}

func (c *CTCoordProtocService) RemoveWorker(ctx context.Context, in *certtranscoordinator.RemoveWorkerRequest) (*certtranscoordinator.WorkerRemovedResponse, error) {
	err := c.cs.RemoveWorker(ctx, in.WorkerCount)
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct coordinator remove worker")
		return nil, err
	}

	return &certtranscoordinator.WorkerRemovedResponse{}, nil
}

func (c *CTCoordProtocService) UpdateDuration(ctx context.Context, in *certtranscoordinator.UpdateDurationRequest) (*certtranscoordinator.DurationUpdatedResponse, error) {
	err := c.cs.UpdateDuration(ctx, in.NewDuration)
	if err != nil {
		log.Error().Err(err).Msg("got error calling ct coordinator update duration")
		return nil, err
	}

	return &certtranscoordinator.DurationUpdatedResponse{}, nil
}
