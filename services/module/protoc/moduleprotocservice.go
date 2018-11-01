package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/protocservices/prototypes"
	context "golang.org/x/net/context"
)

type ModuleProtocService struct {
	module am.ModuleService
}

func New(implementation am.ModuleService) *ModuleProtocService {
	return &ModuleProtocService{module: implementation}
}

func (s *ModuleProtocService) Analyze(ctx context.Context, in *module.AnalyzeRequest) (*module.AnalyzedResponse, error) {
	address, addresses, err := s.module.Analyze(ctx, convert.UserContextToDomain(in.UserContext), convert.AddressToDomain(in.Address))
	protocAddrs := make(map[string]*prototypes.AddressData, len(addresses))
	for k, v := range addresses {
		protocAddrs[k] = convert.DomainToAddress(v)
	}

	return &module.AnalyzedResponse{Original: convert.DomainToAddress(address), Addresses: protocAddrs}, err
}
