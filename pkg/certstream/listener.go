package certstream

type Listener interface {
	Init(closeCh chan struct{}) error
	AddETLD(etld string)
	HasETLD(domain string) (string, bool)
}
