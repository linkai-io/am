package protoc

import (
	"github.com/linkai-io/am/am"
	"github.com/linkai-io/am/pkg/convert"
	"github.com/linkai-io/am/protocservices/module"
	"github.com/linkai-io/am/protocservices/prototypes"
	context "golang.org/x/net/context"
)

type BruteProtocService struct {
	bruteservice am.ModuleService
}

func New(implementation am.ModuleService) *BruteProtocService {
	return &BruteProtocService{bruteservice: implementation}
}

func (s *BruteProtocService) Analyze(ctx context.Context, in *module.AnalyzeRequest) (*module.AnalyzedResponse, error) {
	address, addresses, err := s.bruteservice.Analyze(ctx, convert.UserContextToDomain(in.UserContext), convert.AddressToDomain(in.Address))
	protocAddrs := make(map[string]*prototypes.AddressData, len(addresses))
	for k, v := range addresses {
		protocAddrs[k] = convert.DomainToAddress(v)
	}

	return &module.AnalyzedResponse{Original: convert.DomainToAddress(address), Addresses: protocAddrs}, err
}
