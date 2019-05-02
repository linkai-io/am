package mock

type CertListener struct {
	InitFn      func(closeCh chan struct{}) error
	InitInvoked bool

	AddETLDFn      func(etld string)
	AddETLDInvoked bool

	HasETLDFn      func(domain string) (string, bool)
	HasETLDInvoked bool
}

func (l *CertListener) Init(closeCh chan struct{}) error {
	l.InitInvoked = true
	return l.Init(closeCh)
}

func (l *CertListener) AddETLD(etld string) {
	l.AddETLDInvoked = true
	l.AddETLDFn(etld)
}

func (l *CertListener) HasETLD(domain string) (string, bool) {
	l.HasETLDInvoked = true
	return l.HasETLDFn(domain)
}
