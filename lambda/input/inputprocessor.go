package main

import (
	"context"

	"github.com/linkai-io/am/am"
)

type Processor struct {
	ctx  context.Context
	sg   am.ScanGroupService
	addr am.AddressService
}

func New(sg am.ScanGroupService, addr am.AddressService) *Processor {
	return &Processor{sg: sg, addr: addr}
}

func (p *Processor) Validate(userContext am.UserContext, groupID int, input string) (bool, error) {
	oid, group, err := p.sg.Get(p.ctx, userContext, groupID)
	if err != nil {
		return false, err
	}
	if oid != group.OrgID {
		return false, am.ErrOrgIDMismatch
	}
	return true, nil
}
