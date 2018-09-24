package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/protocservices/prototypes"
	context "golang.org/x/net/context"
)

type NSProtocService struct {
	nsservice am.ModuleService
}

func New(implementation am.ModuleService) *NSProtocService {
	return &NSProtocService{nsservice: implementation}
}

func (s *NSProtocService) Analyze(ctx context.Context, in *module.AnalyzeRequest) (*module.AnalyzedResponse, error) {
	addresses, err := s.nsservice.Analyze(ctx, convert.AddressToDomain(in.Address))
	protocAddrs := make(map[string]*prototypes.AddressData, len(addresses))
	for k, v := range addresses {
		protocAddrs[k] = convert.DomainToAddress(v)
	}

	return &module.AnalyzedResponse{Addresses: protocAddrs}, err
}
