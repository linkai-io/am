package mock

type Storage struct {
	InitFn      func(config []byte) error
	InitInvoked bool

	WriteFn      func(data []byte) (string, string, error)
	WriteInvoked bool
}

func (s *Storage) Init(config []byte) error {
	s.InitInvoked = true
	return s.InitFn(config)
}

func (s *Storage) Write(data []byte) (string, string, error) {
	s.WriteInvoked = true
	return s.WriteFn(data)
}
